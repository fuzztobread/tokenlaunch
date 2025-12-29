package queue

import (
	"context"

	"tokenlaunch/internal/domain"
)

type Publisher interface {
	Publish(ctx context.Context, msg domain.Message) error
	Close() error
}

type Consumer interface {
	Consume(ctx context.Context, handler func(msg domain.Message) error) error
	Close() error
}
