package domain

import "time"

type Message struct {
	ID         string
	ExternalID string
	Author     string
	Username   string
	Content    string
	Source     Source
	CreatedAt  time.Time
}

type Source string

const (
	SourceTwitter  Source = "twitter"
	SourceDiscord  Source = "discord"
	SourceTelegram Source = "telegram"
)
