package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type HTTPSender struct {
	client  *http.Client
	baseURL string
	logger  *slog.Logger
}

func NewHTTPSender(baseURL string, timeout time.Duration, logger *slog.Logger) *HTTPSender {
	return &HTTPSender{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		logger:  logger,
	}
}

func (sender *HTTPSender) SendInviteEmail(ctx context.Context, email string, teamName string) error {
	body := map[string]string{
		"email":     email,
		"team_name": teamName,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal email body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.baseURL+"/post", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create email request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sender.client.Do(req)
	if err != nil {
		return fmt.Errorf("send email request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("email service returned status %d", resp.StatusCode)
	}

	sender.logger.InfoContext(ctx, "invite email sent via http",
		slog.String("email", email),
		slog.String("team_name", teamName),
	)

	return nil
}
