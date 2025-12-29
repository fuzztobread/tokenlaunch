package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tokenlaunch/internal/config"
	"tokenlaunch/internal/scraper"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("starting scraper")
	log.Printf("instance: %s", cfg.Scraper.Instance)
	log.Printf("accounts: %v", cfg.Scraper.Accounts)
	log.Printf("interval: %s", cfg.Scraper.Interval)

	nitter := scraper.NewNitter(cfg.Scraper.Instance)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go run(ctx, cfg, nitter)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("shutting down")
	cancel()
}

func run(ctx context.Context, cfg *config.Config, s scraper.Scraper) {
	ticker := time.NewTicker(cfg.Scraper.Interval)
	defer ticker.Stop()

	// Initial scrape
	scrapeAll(ctx, cfg.Scraper.Accounts, s)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			scrapeAll(ctx, cfg.Scraper.Accounts, s)
		}
	}
}

func scrapeAll(ctx context.Context, accounts []string, s scraper.Scraper) {
	for _, account := range accounts {
		messages, err := s.Scrape(ctx, account)
		if err != nil {
			log.Printf("[ERROR] @%s: %v", account, err)
			continue
		}

		for _, msg := range messages {
			log.Printf("[TWEET] @%s: %s", msg.Username, truncate(msg.Content, 60))
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
