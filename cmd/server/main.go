package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"tokenlaunch/internal/api"
	"tokenlaunch/internal/config"
	"tokenlaunch/internal/storage"
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

	server := api.NewServer(repo)

	go func() {
		log.Printf("server starting on %s", cfg.Server.Port)
		if err := server.Start(cfg.Server.Port); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("shutting down")
	server.Shutdown()
}
