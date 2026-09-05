package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type BrevoEmailSender struct {
	apiKey      string
	senderEmail string
	senderName  string
}

func NewBrevoEmailSender(apiKey, senderEmail, senderName string) *BrevoEmailSender {
	return &BrevoEmailSender{apiKey: apiKey, senderEmail: senderEmail, senderName: senderName}
}

type brevoContact struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type brevoRequest struct {
	Sender      brevoContact   `json:"sender"`
	To          []brevoContact `json:"to"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlContent"`
}

func (s *BrevoEmailSender) Send(ctx context.Context, to, subject, body string) error {
	payload := brevoRequest{
		Sender:      brevoContact{Email: s.senderEmail, Name: s.senderName},
		To:          []brevoContact{{Email: to}},
		Subject:     subject,
		HTMLContent: body,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.brevo.com/v3/smtp/email", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("brevo API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
