package worker

import (
	"bytes"
	"context"
	"html/template"
	"log"

	"tokenlaunch/internal/classifier"
	"tokenlaunch/internal/domain"
	"tokenlaunch/internal/notifier"
	"tokenlaunch/internal/queue"
	"tokenlaunch/internal/storage"
)

type Broadcaster interface {
	Broadcast(msg string)
}

type Consumer struct {
	consumer    queue.Consumer
	repo        storage.MessageRepository
	classifier  classifier.Classifier
	notifier    notifier.Notifier
	broadcaster Broadcaster
	feedTmpl    *template.Template
}

func NewConsumer(c queue.Consumer, r storage.MessageRepository, cl classifier.Classifier, n notifier.Notifier, b Broadcaster) *Consumer {
	tmpl := template.Must(template.New("feed-item").Parse(`
<div class="item {{.Classification}}">
    <div class="item-head">
        <div class="item-author">@{{.Username}}</div>
        <div class="item-time">{{.TimeAgo}}</div>
    </div>
    <div class="item-body">{{.Content}}</div>
    {{if .Classification}}
    <div class="tag {{.Classification}}">{{.Classification}}</div>
    {{end}}
</div>`))

	return &Consumer{
		consumer:    c,
		repo:        r,
		classifier:  cl,
		notifier:    n,
		broadcaster: b,
		feedTmpl:    tmpl,
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
		result = &classifier.Result{Classification: classifier.ClassificationNone}
	}

	view := map[string]string{
		"Username":       msg.Username,
		"Content":        msg.Content,
		"TimeAgo":        "just now",
		"Classification": string(result.Classification),
	}

	if result.Classification == classifier.ClassificationNone {
		view["Classification"] = ""
	}

	var buf bytes.Buffer
	if err := w.feedTmpl.Execute(&buf, view); err == nil {
		w.broadcaster.Broadcast(buf.String())
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
