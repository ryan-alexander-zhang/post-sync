package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/erpang/post-sync/internal/domain"
)

type telegramTargetConfig struct {
	ChatID                string `json:"chatId"`
	TopicID               int64  `json:"topicId,omitempty"`
	TopicName             string `json:"topicName,omitempty"`
	DisableNotification   bool   `json:"disableNotification,omitempty"`
	DisableWebPagePreview bool   `json:"disableWebPagePreview,omitempty"`
}

func normalizeTelegramTarget(
	targetKey string,
	targetType string,
	targetName string,
	rawConfig map[string]any,
) (normalizedTargetKey string, normalizedTargetType string, configJSON string, err error) {
	chatID := strings.TrimSpace(targetKey)
	if chatID == "" {
		chatID = strings.TrimSpace(asString(rawConfig["chatId"]))
	}
	if chatID == "" {
		return "", "", "", fmt.Errorf("%w: telegram group chat id is required", ErrValidation)
	}

	config := telegramTargetConfig{
		ChatID:                chatID,
		DisableNotification:   asBool(rawConfig["disableNotification"]),
		DisableWebPagePreview: asBool(rawConfig["disableWebPagePreview"]),
		TopicName:             strings.TrimSpace(asString(rawConfig["topicName"])),
	}

	if topicID, ok := asInt64(rawConfig["topicId"]); ok {
		config.TopicID = topicID
	}
	if threadID, ok := asInt64(rawConfig["messageThreadId"]); ok && config.TopicID == 0 {
		config.TopicID = threadID
	}

	if config.TopicID > 0 {
		normalizedTargetType = domain.TargetTypeTelegramTopic
		normalizedTargetKey = buildTelegramTargetKey(chatID, config.TopicID)
	} else {
		normalizedTargetType = domain.TargetTypeTelegramGrp
		normalizedTargetKey = chatID
	}

	if strings.TrimSpace(targetType) != "" {
		switch targetType {
		case domain.TargetTypeTelegramGrp:
			if config.TopicID > 0 {
				return "", "", "", fmt.Errorf("%w: telegram_group cannot have topicId", ErrValidation)
			}
		case domain.TargetTypeTelegramTopic:
			if config.TopicID <= 0 {
				return "", "", "", fmt.Errorf("%w: telegram_topic requires topicId", ErrValidation)
			}
		default:
			return "", "", "", fmt.Errorf("%w: unsupported telegram target type", ErrValidation)
		}
		normalizedTargetType = targetType
	}

	if normalizedTargetType == domain.TargetTypeTelegramTopic && strings.TrimSpace(targetName) == "" {
		return "", "", "", fmt.Errorf("%w: topic target requires targetName", ErrValidation)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return "", "", "", fmt.Errorf("%w: invalid telegram config", ErrValidation)
	}

	return normalizedTargetKey, normalizedTargetType, string(data), nil
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
