package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"tokenlaunch/internal/classifier"
	"tokenlaunch/internal/config"
	"tokenlaunch/internal/notifier"
	"tokenlaunch/internal/queue"
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

	consumer, err := queue.NewKafkaConsumer(cfg.Queue.Brokers, cfg.Queue.GroupID, cfg.Queue.Topic)
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}
	defer consumer.Close()

	cl := classifier.NewOpenRouter(cfg.Classifier.APIKey, cfg.Classifier.Model)
	nt := notifier.NewTelegram(cfg.Notifier.TelegramToken, cfg.Notifier.TelegramChatIDs)

	w := worker.NewConsumer(consumer, repo, cl, nt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := w.Start(ctx); err != nil {
			log.Printf("consumer error: %v", err)
		}
	}()

	log.Printf("consumer started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("shutting down")
	cancel()
}
