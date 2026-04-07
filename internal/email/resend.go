package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ResendSender struct {
	apiKey string
	from   string
}

func NewResendSender(apiKey, from string) *ResendSender {
	return &ResendSender{apiKey: apiKey, from: from}
}

func (r *ResendSender) Send(to, subject string, data ForwardData) error {
	var body bytes.Buffer
	if err := emailTemplate.Execute(&body, data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	payload, err := json.Marshal(map[string]any{
		"from":    r.from,
		"to":      []string{to},
		"subject": subject,
		"html":    body.String(),
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error (%d): %s", resp.StatusCode, respBody)
	}

	return nil
}
