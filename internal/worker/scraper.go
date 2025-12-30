package worker

import (
	"context"
	"log"
	"time"

	"tokenlaunch/internal/config"
	"tokenlaunch/internal/queue"
	"tokenlaunch/internal/scraper"
)

type Scraper struct {
	scraper   scraper.Scraper
	publisher queue.Publisher
	accounts  []string
	interval  time.Duration
	seen      map[string]bool
}

func NewScraper(s scraper.Scraper, p queue.Publisher, cfg config.ScraperConfig) *Scraper {
	return &Scraper{
		scraper:   s,
		publisher: p,
		accounts:  cfg.Accounts,
		interval:  cfg.Interval,
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
	for _, account := range w.accounts {
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
