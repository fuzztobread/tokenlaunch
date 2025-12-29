package notifier

import (
	"context"

	"tokenlaunch/internal/classifier"
	"tokenlaunch/internal/domain"
)

type Notification struct {
	Message domain.Message
	Result  classifier.Result
}

type Notifier interface {
	Notify(ctx context.Context, n Notification) error
}
