package service

import (
	"strings"
	"testing"

	"github.com/erpang/post-sync/internal/domain"
)

func TestNormalizeTelegramTargetBuildsTopicTarget(t *testing.T) {
	targetKey, targetType, configJSON, err := normalizeTelegramTarget(
		"-100123",
		"",
		"Announcements",
		map[string]any{
			"topicId":               "42",
			"topicName":             "Announcements",
			"disableNotification":   true,
			"disableWebPagePreview": true,
		},
	)
	if err != nil {
		t.Fatalf("normalizeTelegramTarget() error = %v", err)
	}

	if targetKey != "-100123:topic:42" {
		t.Fatalf("targetKey = %q", targetKey)
	}
	if targetType != domain.TargetTypeTelegramTopic {
		t.Fatalf("targetType = %q", targetType)
	}
	if !strings.Contains(configJSON, `"chatId":"-100123"`) {
		t.Fatalf("configJSON = %q", configJSON)
	}
}

func TestNormalizeTelegramTargetBuildsGroupTarget(t *testing.T) {
	targetKey, targetType, _, err := normalizeTelegramTarget(
		"-100123",
		"",
		"Main Group",
		map[string]any{},
	)
	if err != nil {
		t.Fatalf("normalizeTelegramTarget() error = %v", err)
	}

	if targetKey != "-100123" {
		t.Fatalf("targetKey = %q", targetKey)
	}
	if targetType != domain.TargetTypeTelegramGrp {
		t.Fatalf("targetType = %q", targetType)
	}
}
