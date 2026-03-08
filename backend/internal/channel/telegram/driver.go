package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/erpang/post-sync/internal/channel"
	"github.com/erpang/post-sync/internal/domain"
)

type Driver struct{}

func New() *Driver {
	return &Driver{}
}

func (d *Driver) Type() string {
	return domain.ChannelTypeTelegram
}

func (d *Driver) ValidateAccount(config map[string]any, secretRef string) error {
	if strings.TrimSpace(secretRef) == "" {
		return fmt.Errorf("telegram secretRef is required")
	}
	return nil
}

func (d *Driver) ValidateTarget(config map[string]any, targetKey string) error {
	if strings.TrimSpace(targetKey) == "" {
		return fmt.Errorf("telegram chat id is required")
	}
	return nil
}

func (d *Driver) Render(input channel.RenderInput) (channel.RenderedMessage, error) {
	return channel.RenderedMessage{
		Title:      input.ContentTitle,
		Body:       strings.TrimSpace(input.ContentBody),
		RenderMode: domain.RenderModeTelegram,
	}, nil
}

func (d *Driver) Send(ctx context.Context, request channel.SendRequest) (channel.SendResult, error) {
	if strings.TrimSpace(request.SecretValue) == "" {
		return channel.SendResult{}, fmt.Errorf("telegram secret value is empty")
	}

	baseURL := "https://api.telegram.org"
	if override, ok := request.Config["apiBaseURL"].(string); ok && strings.TrimSpace(override) != "" {
		baseURL = strings.TrimSpace(override)
	}

	payload := map[string]any{
		"chat_id":                  request.TargetKey,
		"text":                     request.Body,
		"parse_mode":               "HTML",
		"disable_web_page_preview": false,
	}

	if value, ok := request.Config["disableNotification"].(bool); ok {
		payload["disable_notification"] = value
	}
	if value, ok := request.Config["disableWebPagePreview"].(bool); ok {
		payload["disable_web_page_preview"] = value
	}
	if value, ok := request.Config["messageThreadId"].(float64); ok {
		payload["message_thread_id"] = int(value)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return channel.SendResult{}, err
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", strings.TrimRight(baseURL, "/"), url.PathEscape(request.SecretValue))
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return channel.SendResult{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return channel.SendResult{}, err
	}
	defer response.Body.Close()

	var telegramResponse struct {
		OK          bool            `json:"ok"`
		Description string          `json:"description"`
		Result      json.RawMessage `json:"result"`
	}

	if err := json.NewDecoder(response.Body).Decode(&telegramResponse); err != nil {
		return channel.SendResult{}, err
	}
	if !telegramResponse.OK || response.StatusCode >= 400 {
		return channel.SendResult{}, fmt.Errorf("telegram send failed: %s", telegramResponse.Description)
	}

	return channel.SendResult{
		ExternalMessageID: extractMessageID(telegramResponse.Result),
		ProviderResponse:  string(telegramResponse.Result),
	}, nil
}

func extractMessageID(raw json.RawMessage) string {
	var payload struct {
		MessageID int64 `json:"message_id"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	if payload.MessageID == 0 {
		return ""
	}
	return fmt.Sprintf("%d", payload.MessageID)
}
