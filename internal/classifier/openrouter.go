package classifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"tokenlaunch/internal/domain"
)

type OpenRouter struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenRouter(apiKey, model string) *OpenRouter {
	return &OpenRouter{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (o *OpenRouter) Classify(ctx context.Context, msg domain.Message) (*Result, error) {
	prompt := fmt.Sprintf(`Analyze this tweet and classify it:

Tweet by @%s:
"%s"

Classify as one of:
- "launch": Announces a new crypto token launch
- "endorsement": Promotes or endorses an existing crypto token
- "none": Not related to crypto tokens

Respond in JSON format only:
{
  "classification": "launch|endorsement|none",
  "token": "token symbol if mentioned, empty otherwise",
  "confidence": 0.0-1.0,
  "reason": "brief explanation"
}`, msg.Username, msg.Content)

	reqBody := map[string]any{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	return parseResponse(apiResp.Choices[0].Message.Content)
}

func parseResponse(content string) (*Result, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result struct {
		Classification string  `json:"classification"`
		Token          string  `json:"token"`
		Confidence     float64 `json:"confidence"`
		Reason         string  `json:"reason"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return &Result{Classification: ClassificationNone}, nil
	}

	classification := ClassificationNone
	switch result.Classification {
	case "launch":
		classification = ClassificationLaunch
	case "endorsement":
		classification = ClassificationEndorsement
	}

	return &Result{
		Classification: classification,
		Token:          result.Token,
		Confidence:     result.Confidence,
		Reason:         result.Reason,
	}, nil
}
