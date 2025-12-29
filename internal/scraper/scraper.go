package scraper

import (
	"context"

	"tokenlaunch/internal/domain"
)

type Scraper interface {
	Scrape(ctx context.Context, account string) ([]domain.Message, error)
}
