package worker

import (
	"bytes"
	"context"
	"fmt"
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

	// Save to DB
	if err := w.repo.Save(ctx, msg); err != nil {
		log.Printf("[DB ERROR] save failed: %v", err)
		return err
	}
	log.Printf("[DB] saved message id=%s", msg.ID)

	// Classify with LLM
	log.Printf("[CLASSIFY] sending to LLM...")
	result, err := w.classifier.Classify(ctx, msg)
	if err != nil {
		log.Printf("[CLASSIFY ERROR] %v", err)
		result = &classifier.Result{Classification: classifier.ClassificationNone}
	} else {
		log.Printf("[CLASSIFY] result: type=%s, token=%s, confidence=%.2f, reason=%s",
			result.Classification, result.Token, result.Confidence, truncate(result.Reason, 50))
	}

	// Save classification to DB
	if result.Classification != classifier.ClassificationNone {
		w.repo.UpdateClassification(ctx, msg.ID, string(result.Classification), result.Token, result.Confidence)
	}

	// Broadcast to SSE
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
		log.Printf("[SSE] broadcasted to dashboard")
	}

	// Notify if launch/endorsement
	if result.Classification != classifier.ClassificationNone {
		log.Printf("[ALERT] %s detected! token=%s", result.Classification, result.Token)

		// Broadcast toast notification
		toast := fmt.Sprintf(`<div id="toast" class="toast show" hx-swap-oob="true">%s detected: %s</div>`,
			result.Classification, result.Token)
		w.broadcaster.Broadcast(toast)

		if err := w.notifier.Notify(ctx, notifier.Notification{
			Message: msg,
			Result:  *result,
		}); err != nil {
			log.Printf("[NOTIFY ERROR] telegram: %v", err)
		} else {
			log.Printf("[NOTIFY] sent to telegram")
		}
	}

	log.Printf("[DONE] processed message id=%s", msg.ID)
	return nil
}
