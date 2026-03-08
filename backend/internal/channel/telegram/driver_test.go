package telegram

import (
	"testing"

	"github.com/erpang/post-sync/internal/channel"
	"github.com/erpang/post-sync/internal/domain"
)

func TestNormalizeTargetBuildsTopicTarget(t *testing.T) {
	driver := New()

	normalized, err := driver.NormalizeTarget(channel.TargetInput{
		TargetKey:  "-100123",
		TargetName: "Announcements",
		Config: map[string]any{
			"topicId":               "42",
			"topicName":             "Announcements",
			"disableNotification":   true,
			"disableWebPagePreview": true,
		},
	})
	if err != nil {
		t.Fatalf("NormalizeTarget() error = %v", err)
	}

	if normalized.TargetKey != "-100123:topic:42" {
		t.Fatalf("TargetKey = %q", normalized.TargetKey)
	}
	if normalized.TargetType != domain.TargetTypeTelegramTopic {
		t.Fatalf("TargetType = %q", normalized.TargetType)
	}
	if normalized.Config["chatId"] != "-100123" {
		t.Fatalf("Config chatId = %v", normalized.Config["chatId"])
	}
}

func TestNormalizeTargetBuildsGroupTarget(t *testing.T) {
	driver := New()

	normalized, err := driver.NormalizeTarget(channel.TargetInput{
		TargetKey:  "-100123",
		TargetName: "Main Group",
		Config:     map[string]any{},
	})
	if err != nil {
		t.Fatalf("NormalizeTarget() error = %v", err)
	}

	if normalized.TargetKey != "-100123" {
		t.Fatalf("TargetKey = %q", normalized.TargetKey)
	}
	if normalized.TargetType != domain.TargetTypeTelegramGrp {
		t.Fatalf("TargetType = %q", normalized.TargetType)
	}
}
