package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"tokenlaunch/internal/api"
	"tokenlaunch/internal/classifier"
	"tokenlaunch/internal/config"
	"tokenlaunch/internal/notifier"
	"tokenlaunch/internal/queue"
	"tokenlaunch/internal/redis"
	"tokenlaunch/internal/storage"
	"tokenlaunch/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	repo, err := storage.NewPostgres(cfg.Storage.DSN)
	if err != nil {
		log.Fatalf("failed to connect to storage: %v", err)
	}
	defer repo.Close()

	rdb, err := redis.New(cfg.Redis.Addr)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	consumer, err := queue.NewKafkaConsumer(cfg.Queue.Brokers, cfg.Queue.GroupID, cfg.Queue.Topic)
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}
	defer consumer.Close()

	cl := classifier.NewOpenRouter(cfg.Classifier.APIKey, cfg.Classifier.Model)
	nt := notifier.NewTelegram(cfg.Notifier.TelegramToken, cfg.Notifier.TelegramChatIDs)

	server := api.NewServer(repo, rdb)

	w := worker.NewConsumer(consumer, repo, cl, nt, server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := w.Start(ctx); err != nil {
			log.Printf("consumer error: %v", err)
		}
	}()

	go func() {
		log.Printf("server starting on %s", cfg.Server.Port)
		if err := server.Start(cfg.Server.Port); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	log.Printf("app started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("shutting down")
	cancel()
	server.Shutdown()
}
