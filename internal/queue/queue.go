package queue

import (
	"context"

	"tokenlaunch/internal/domain"
)

type Publisher interface {
	Publish(ctx context.Context, msg domain.Message) error
	Close() error
}
