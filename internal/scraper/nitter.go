package scraper

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"

	"tokenlaunch/internal/domain"
)

type Nitter struct {
	instance string
	client   *http.Client
	parser   *gofeed.Parser
}

func NewNitter(instance string) *Nitter {
	return &Nitter{
		instance: instance,
		client:   &http.Client{Timeout: 15 * time.Second},
		parser:   gofeed.NewParser(),
	}
}

func (n *Nitter) Scrape(ctx context.Context, account string) ([]domain.Message, error) {
	url := fmt.Sprintf("https://%s/%s/rss", n.instance, account)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "curl/8.0")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml, */*")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	feed, err := n.parser.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	messages := make([]domain.Message, 0, len(feed.Items))
	for _, item := range feed.Items {
		createdAt := time.Now()
		if item.PublishedParsed != nil {
			createdAt = *item.PublishedParsed
		}

		messages = append(messages, domain.Message{
			ID:         generateID(item.GUID),
			ExternalID: item.GUID,
			Author:     feed.Title,
			Username:   account,
			Content:    item.Title,
			Source:     domain.SourceTwitter,
			CreatedAt:  createdAt,
		})
	}

	return messages, nil
}

func generateID(guid string) string {
	hash := md5.Sum([]byte(guid))
	return fmt.Sprintf("%x", hash)[:12]
}
