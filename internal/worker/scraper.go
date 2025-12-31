package worker

import (
	"context"
	"log"
	"time"

	"tokenlaunch/internal/queue"
	"tokenlaunch/internal/redis"
	"tokenlaunch/internal/scraper"
)

type Scraper struct {
	scraper   scraper.Scraper
	publisher queue.Publisher
	redis     *redis.Client
	instance  string
	interval  time.Duration
	seen      map[string]bool
}

func NewScraper(s scraper.Scraper, p queue.Publisher, r *redis.Client, instance string, interval time.Duration) *Scraper {
	return &Scraper{
		scraper:   s,
		publisher: p,
		redis:     r,
		instance:  instance,
		interval:  interval,
		seen:      make(map[string]bool),
	}
}

func (w *Scraper) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.scrapeAll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.scrapeAll(ctx)
		}
	}
}

func (w *Scraper) scrapeAll(ctx context.Context) {
	accounts, err := w.redis.GetAccounts(ctx)
	if err != nil {
		log.Printf("[ERROR] failed to get accounts from redis: %v", err)
		return
	}

	if len(accounts) == 0 {
		log.Printf("[SCRAPE] no accounts configured")
		return
	}

	for _, account := range accounts {
		messages, err := w.scraper.Scrape(ctx, account)
		if err != nil {
			log.Printf("[ERROR] @%s: %v", account, err)
			continue
		}

		log.Printf("[SCRAPE] @%s: fetched %d tweets", account, len(messages))

		newCount := 0
		dupCount := 0

		for _, msg := range messages {
			if w.seen[msg.ID] {
				dupCount++
				continue
			}
			w.seen[msg.ID] = true
			newCount++

			if err := w.publisher.Publish(ctx, msg); err != nil {
				log.Printf("[ERROR] publish: %v", err)
				continue
			}
			log.Printf("[QUEUED] @%s: %s", msg.Username, truncate(msg.Content, 60))
		}

		log.Printf("[STATS] @%s: new=%d, duplicates=%d, seen_total=%d", account, newCount, dupCount, len(w.seen))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
