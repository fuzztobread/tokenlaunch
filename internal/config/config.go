package config

import (
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server     ServerConfig
	Scraper    ScraperConfig
	Redis      RedisConfig
	Queue      QueueConfig
	Storage    StorageConfig
	Classifier ClassifierConfig
	Notifier   NotifierConfig
}

type ServerConfig struct {
	Port string
}

type ScraperConfig struct {
	Instance string
	Interval time.Duration
}

type RedisConfig struct {
	Addr string
}

type QueueConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

type StorageConfig struct {
	DSN string
}

type ClassifierConfig struct {
	APIKey string
	Model  string
}

type NotifierConfig struct {
	TelegramToken   string
	TelegramChatIDs []string
}

func Load() (*Config, error) {
	k := koanf.New(".")

	envPath := os.Getenv("ENV_PATH")
	if envPath == "" {
		envPath = ".env"
	}

	parser := dotenv.ParserEnv("", "_", strings.ToLower)

	if err := k.Load(file.Provider(envPath), parser); err != nil {
		return nil, err
	}

	cfg := &Config{}

	cfg.Server.Port = k.String("server.port")

	cfg.Scraper.Instance = k.String("scraper.instance")
	cfg.Scraper.Interval = k.Duration("scraper.interval")

	cfg.Redis.Addr = k.String("redis.addr")

	cfg.Queue.Brokers = strings.Split(k.String("queue.brokers"), ",")
	cfg.Queue.Topic = k.String("queue.topic")
	cfg.Queue.GroupID = k.String("queue.group.id")

	cfg.Storage.DSN = k.String("storage.dsn")

	cfg.Classifier.APIKey = k.String("classifier.api.key")
	cfg.Classifier.Model = k.String("classifier.model")

	cfg.Notifier.TelegramToken = k.String("notifier.telegram.token")
	cfg.Notifier.TelegramChatIDs = strings.Split(k.String("notifier.telegram.chat.ids"), ",")

	return cfg, nil
}
