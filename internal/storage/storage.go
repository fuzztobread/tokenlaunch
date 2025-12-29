package storage

import (
	"context"

	"tokenlaunch/internal/domain"
)

type MessageRepository interface {
	Save(ctx context.Context, msg domain.Message) error
	FindByID(ctx context.Context, id string) (*domain.Message, error)
	FindAll(ctx context.Context, limit, offset int) ([]domain.Message, error)
	Exists(ctx context.Context, id string) (bool, error)
}
