package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Telegram struct {
	botToken string
	chatIDs  []string
	client   *http.Client
}

func NewTelegram(botToken string, chatIDs []string) *Telegram {
	return &Telegram{
		botToken: botToken,
		chatIDs:  chatIDs,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *Telegram) Notify(ctx context.Context, n Notification) error {
	text := formatMessage(n)

	for _, chatID := range t.chatIDs {
		if err := t.send(ctx, chatID, text); err != nil {
			return err
		}
	}

	return nil
}

func (t *Telegram) send(ctx context.Context, chatID, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	body, _ := json.Marshal(map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram error: %d", resp.StatusCode)
	}

	return nil
}

func formatMessage(n Notification) string {
	icon := "📢"
	if n.Result.Classification == "launch" {
		icon = "🚀"
	}

	return fmt.Sprintf(`%s <b>%s detected</b>

<b>Author:</b> @%s
<b>Token:</b> %s
<b>Confidence:</b> %.0f%%

<b>Tweet:</b>
%s

<b>Reason:</b> %s`,
		icon,
		n.Result.Classification,
		n.Message.Username,
		n.Result.Token,
		n.Result.Confidence*100,
		n.Message.Content,
		n.Result.Reason,
	)
}
