package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
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

func (d *Driver) ValidateAccount(input channel.AccountValidationInput) error {
	secretRef := strings.TrimSpace(input.SecretRef)
	if secretRef == "" {
		return fmt.Errorf("telegram secretRef is required")
	}
	if _, ok := os.LookupEnv(secretRef); !ok {
		return fmt.Errorf("environment variable %s is not set", secretRef)
	}
	return nil
}

func (d *Driver) NormalizeTarget(input channel.TargetInput) (channel.NormalizedTarget, error) {
	chatID := strings.TrimSpace(input.TargetKey)
	if value, ok := input.Config["chatId"].(string); ok && strings.TrimSpace(value) != "" {
		chatID = strings.TrimSpace(value)
	}
	if chatID == "" {
		return channel.NormalizedTarget{}, fmt.Errorf("telegram group chat id is required")
	}

	config := map[string]any{
		"chatId":                chatID,
		"disableNotification":   asBool(input.Config["disableNotification"]),
		"disableWebPagePreview": asBool(input.Config["disableWebPagePreview"]),
	}

	if topicName := strings.TrimSpace(asString(input.Config["topicName"])); topicName != "" {
		config["topicName"] = topicName
	}

	var topicID int64
	if value, ok := asInt64(input.Config["topicId"]); ok {
		topicID = value
	}
	if value, ok := asInt64(input.Config["messageThreadId"]); ok && topicID == 0 {
		topicID = value
	}

	targetType := input.TargetType
	targetKey := chatID

	if topicID > 0 {
		config["topicId"] = topicID
		targetType = domain.TargetTypeTelegramTopic
		targetKey = buildTelegramTargetKey(chatID, topicID)
	}
	if strings.TrimSpace(targetType) == "" {
		targetType = domain.TargetTypeTelegramGrp
	}

	switch targetType {
	case domain.TargetTypeTelegramGrp:
		if topicID > 0 {
			return channel.NormalizedTarget{}, fmt.Errorf("telegram_group cannot have topicId")
		}
		targetKey = chatID
	case domain.TargetTypeTelegramTopic:
		if topicID <= 0 {
			return channel.NormalizedTarget{}, fmt.Errorf("telegram_topic requires topicId")
		}
		targetKey = buildTelegramTargetKey(chatID, topicID)
	default:
		return channel.NormalizedTarget{}, fmt.Errorf("unsupported telegram target type")
	}

	if targetType == domain.TargetTypeTelegramTopic && strings.TrimSpace(input.TargetName) == "" {
		return channel.NormalizedTarget{}, fmt.Errorf("topic target requires targetName")
	}

	return channel.NormalizedTarget{
		TargetType: targetType,
		TargetKey:  targetKey,
		Config:     config,
	}, nil
}

func (d *Driver) Render(input channel.RenderInput) (channel.RenderedMessage, error) {
	return channel.RenderedMessage{
		Title:      input.ContentTitle,
		Body:       strings.TrimSpace(input.ContentBody),
		RenderMode: domain.RenderModeTelegram,
	}, nil
}

func (d *Driver) Send(ctx context.Context, request channel.SendRequest) (channel.SendResult, error) {
	secretValue := strings.TrimSpace(os.Getenv(request.Account.SecretRef))
	if secretValue == "" {
		return channel.SendResult{}, fmt.Errorf("telegram secret value is empty")
	}

	baseURL := "https://api.telegram.org"
	if override, ok := request.Target.Config["apiBaseURL"].(string); ok && strings.TrimSpace(override) != "" {
		baseURL = strings.TrimSpace(override)
	}

	payload := map[string]any{
		"chat_id":                  request.Target.Key,
		"text":                     request.Body,
		"parse_mode":               "HTML",
		"disable_web_page_preview": false,
	}
	if value, ok := request.Target.Config["chatId"].(string); ok && strings.TrimSpace(value) != "" {
		payload["chat_id"] = strings.TrimSpace(value)
	}

	if value, ok := request.Target.Config["disableNotification"].(bool); ok {
		payload["disable_notification"] = value
	}
	if value, ok := request.Target.Config["disableWebPagePreview"].(bool); ok {
		payload["disable_web_page_preview"] = value
	}
	if topicID, ok := asInt64(request.Target.Config["messageThreadId"]); ok && topicID > 0 {
		payload["message_thread_id"] = topicID
	}
	if topicID, ok := asInt64(request.Target.Config["topicId"]); ok && topicID > 0 {
		payload["message_thread_id"] = topicID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return channel.SendResult{}, err
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", strings.TrimRight(baseURL, "/"), url.PathEscape(secretValue))
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

func buildTelegramTargetKey(chatID string, topicID int64) string {
	if topicID <= 0 {
		return chatID
	}
	return fmt.Sprintf("%s:topic:%d", chatID, topicID)
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func asBool(value any) bool {
	typed, ok := value.(bool)
	return ok && typed
}

func asInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int64:
		return typed, true
	case float64:
		return int64(typed), true
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return 0, false
		}
		parsed, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
