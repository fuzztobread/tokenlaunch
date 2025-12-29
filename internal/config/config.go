package config

import (
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server     ServerConfig     `koanf:"server"`
	Scraper    ScraperConfig    `koanf:"scraper"`
	Queue      QueueConfig      `koanf:"queue"`
	Storage    StorageConfig    `koanf:"storage"`
	Classifier ClassifierConfig `koanf:"classifier"`
	Notifier   NotifierConfig   `koanf:"notifier"`
}

type ServerConfig struct {
	Port string `koanf:"port"`
}

type ScraperConfig struct {
	Instance string        `koanf:"instance"`
	Accounts []string      `koanf:"accounts"`
	Interval time.Duration `koanf:"interval"`
}

type QueueConfig struct {
	Brokers []string `koanf:"brokers"`
	Topic   string   `koanf:"topic"`
	GroupID string   `koanf:"group_id"`
}

type StorageConfig struct {
	DSN string `koanf:"dsn"`
}

type ClassifierConfig struct {
	APIKey string `koanf:"api_key"`
	Model  string `koanf:"model"`
}

type NotifierConfig struct {
	TelegramToken   string   `koanf:"telegram_token"`
	TelegramChatIDs []string `koanf:"telegram_chat_ids"`
}

func Load() (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
