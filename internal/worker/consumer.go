package worker

import (
	"context"
	"log"

	"tokenlaunch/internal/classifier"
	"tokenlaunch/internal/domain"
	"tokenlaunch/internal/notifier"
	"tokenlaunch/internal/queue"
	"tokenlaunch/internal/storage"
)

type Consumer struct {
	consumer   queue.Consumer
	repo       storage.MessageRepository
	classifier classifier.Classifier
	notifier   notifier.Notifier
}

func NewConsumer(c queue.Consumer, r storage.MessageRepository, cl classifier.Classifier, n notifier.Notifier) *Consumer {
	return &Consumer{
		consumer:   c,
		repo:       r,
		classifier: cl,
		notifier:   n,
	}
}

func (w *Consumer) Start(ctx context.Context) error {
	return w.consumer.Consume(ctx, w.handleMessage)
}

func (w *Consumer) handleMessage(msg domain.Message) error {
	ctx := context.Background()

	log.Printf("[RECEIVED] @%s: %s", msg.Username, truncate(msg.Content, 60))

	if err := w.repo.Save(ctx, msg); err != nil {
		log.Printf("[ERROR] save: %v", err)
		return err
	}

	result, err := w.classifier.Classify(ctx, msg)
	if err != nil {
		log.Printf("[ERROR] classify: %v", err)
		return nil
	}

	if result.Classification != classifier.ClassificationNone {
		log.Printf("[DETECTED] %s: %s (token: %s, confidence: %.2f)",
			result.Classification, result.Reason, result.Token, result.Confidence)

		if err := w.notifier.Notify(ctx, notifier.Notification{
			Message: msg,
			Result:  *result,
		}); err != nil {
			log.Printf("[ERROR] notify: %v", err)
		}
	}

	return nil
}
