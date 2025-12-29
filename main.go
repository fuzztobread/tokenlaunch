package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type Config struct {
	NitterInstance string
	Accounts       []string
	Interval       time.Duration
}

func main() {
	cfg := Config{
		NitterInstance: getEnv("NITTER_INSTANCE", "nitter.privacyredirect.com"),
		Accounts:       strings.Split(getEnv("ACCOUNTS", "elonmusk"), ","),
		Interval:       30 * time.Second,
	}

	log.Printf("TokenLaunch Scraper starting")
	log.Printf("Instance: %s", cfg.NitterInstance)
	log.Printf("Accounts: %v", cfg.Accounts)

	scrape(cfg)

	ticker := time.NewTicker(cfg.Interval)
	for range ticker.C {
		scrape(cfg)
	}
}

func scrape(cfg Config) {
	for _, account := range cfg.Accounts {
		tweets, err := fetchTweets(cfg.NitterInstance, account)
		if err != nil {
			log.Printf("[ERROR] @%s: %v", account, err)
			continue
		}

		for _, tweet := range tweets {
			log.Printf("[TWEET] @%s: %s", account, truncate(tweet.Title, 60))
		}
	}
}

func fetchTweets(instance, username string) ([]*gofeed.Item, error) {
	url := fmt.Sprintf("https://%s/%s/rss", instance, username)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "curl/8.0")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml, */*")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return feed.Items, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
