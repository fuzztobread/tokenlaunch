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
}

func NewScraper(s scraper.Scraper, p queue.Publisher, cfg config.ScraperConfig) *Scraper {
	return &Scraper{
		scraper:   s,
		publisher: p,
		accounts:  cfg.Accounts,
		interval:  cfg.Interval,
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

		for _, msg := range messages {
			if err := w.publisher.Publish(ctx, msg); err != nil {
				log.Printf("[ERROR] publish: %v", err)
				continue
			}
			log.Printf("[QUEUED] @%s: %s", msg.Username, truncate(msg.Content, 60))
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
