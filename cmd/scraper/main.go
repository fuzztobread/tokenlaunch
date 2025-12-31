package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"tokenlaunch/internal/config"
	"tokenlaunch/internal/queue"
	"tokenlaunch/internal/redis"
	"tokenlaunch/internal/scraper"
	"tokenlaunch/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	rdb, err := redis.New(cfg.Redis.Addr)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	publisher, err := queue.NewKafka(cfg.Queue.Brokers, cfg.Queue.Topic)
	if err != nil {
		log.Fatalf("failed to create queue: %v", err)
	}
	defer publisher.Close()

	nitter := scraper.NewNitter(cfg.Scraper.Instance)

	w := worker.NewScraper(nitter, publisher, rdb, cfg.Scraper.Instance, cfg.Scraper.Interval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Start(ctx)

	log.Printf("scraper started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("shutting down")
	cancel()
}
