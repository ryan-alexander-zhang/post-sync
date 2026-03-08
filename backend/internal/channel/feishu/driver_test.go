package feishu

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/erpang/post-sync/internal/channel"
	"github.com/erpang/post-sync/internal/domain"
)

func TestNormalizeTargetBuildsChatTarget(t *testing.T) {
	driver := New(nil, nil)

	normalized, err := driver.NormalizeTarget(channel.TargetInput{
		TargetKey:  "oc_123",
		TargetName: "Team Chat",
		Config:     map[string]any{},
	})
	if err != nil {
		t.Fatalf("NormalizeTarget() error = %v", err)
	}

	if normalized.TargetType != domain.TargetTypeFeishuChat {
		t.Fatalf("TargetType = %q", normalized.TargetType)
	}
	if normalized.TargetKey != "oc_123" {
		t.Fatalf("TargetKey = %q", normalized.TargetKey)
	}
}

func TestSendUsesTenantAccessTokenFlow(t *testing.T) {
	t.Setenv("FEISHU_APP_ID", "cli_test")
	t.Setenv("FEISHU_APP_SECRET", "secret_test")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/auth/v3/tenant_access_token/internal":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":0,"msg":"ok","tenant_access_token":"t-token","expire":7200}`))
		case "/open-apis/im/v1/messages":
			if got := r.Header.Get("Authorization"); got != "Bearer t-token" {
				t.Fatalf("Authorization = %q", got)
			}

			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if payload["receive_id"] != "oc_123" {
				t.Fatalf("receive_id = %v", payload["receive_id"])
			}
			if payload["msg_type"] != "post" {
				t.Fatalf("msg_type = %v", payload["msg_type"])
			}
			content, _ := payload["content"].(string)
			if !strings.Contains(content, "\"tag\":\"md\"") || !strings.Contains(content, "hello") {
				t.Fatalf("content = %q", content)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":0,"msg":"ok","data":{"message_id":"om_123"}}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := server.Client()
	driver := New(client, NewTokenProvider(client))

	result, err := driver.Send(context.Background(), channel.SendRequest{
		Account: channel.Account{
			Type:      domain.ChannelTypeFeishu,
			SecretRef: "FEISHU_APP_SECRET",
			Config: map[string]any{
				"appIdEnv": "FEISHU_APP_ID",
				"baseUrl":  server.URL,
			},
		},
		Target: channel.Target{
			Type: domain.TargetTypeFeishuChat,
			Key:  "oc_123",
			Config: map[string]any{
				"receiveIdType": "chat_id",
				"chatId":        "oc_123",
			},
		},
		Body:           "hello",
		IdempotencyKey: "dedup-1",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result.ExternalMessageID != "om_123" {
		t.Fatalf("ExternalMessageID = %q", result.ExternalMessageID)
	}
}

func TestValidateAccountAcceptsStaticTokenEnv(t *testing.T) {
	t.Setenv("FEISHU_TENANT_ACCESS_TOKEN", "t-static")

	driver := New(nil, nil)
	err := driver.ValidateAccount(channel.AccountValidationInput{
		SecretRef: "FEISHU_APP_SECRET",
		Config: map[string]any{
			"tokenEnv": "FEISHU_TENANT_ACCESS_TOKEN",
		},
	})
	if err != nil {
		t.Fatalf("ValidateAccount() error = %v", err)
	}
}

func TestTokenProviderUsesStaticTokenEnv(t *testing.T) {
	t.Setenv("FEISHU_TENANT_ACCESS_TOKEN", "t-static")

	provider := NewTokenProvider(nil)
	token, err := provider.GetTenantAccessToken(context.Background(), channel.Account{
		Config: map[string]any{
			"tokenEnv": "FEISHU_TENANT_ACCESS_TOKEN",
		},
	})
	if err != nil {
		t.Fatalf("GetTenantAccessToken() error = %v", err)
	}
	if token != "t-static" {
		t.Fatalf("token = %q", token)
	}
}

func TestResolveConfiguredValueReadsEnv(t *testing.T) {
	t.Setenv("FEISHU_APP_ID", "cli_test")

	value, err := resolveConfiguredValue(map[string]any{"appIdEnv": "FEISHU_APP_ID"}, "appIdEnv", "appId")
	if err != nil {
		t.Fatalf("resolveConfiguredValue() error = %v", err)
	}
	if value != "cli_test" {
		t.Fatalf("value = %q", value)
	}
}

func TestBuildPostContentIncludesOptionalTitle(t *testing.T) {
	content := buildPostContent("Test Title", "# heading\n\nbody")

	payload, ok := content["zh_cn"].(map[string]any)
	if !ok {
		t.Fatalf("zh_cn payload missing")
	}
	if payload["title"] != "Test Title" {
		t.Fatalf("title = %v", payload["title"])
	}

	paragraphs, ok := payload["content"].([][]map[string]string)
	if !ok || len(paragraphs) != 1 || len(paragraphs[0]) != 1 {
		t.Fatalf("content paragraphs malformed: %#v", payload["content"])
	}
	if paragraphs[0][0]["tag"] != "md" {
		t.Fatalf("tag = %q", paragraphs[0][0]["tag"])
	}
	if paragraphs[0][0]["text"] != "# heading\n\nbody" {
		t.Fatalf("text = %q", paragraphs[0][0]["text"])
	}
}

func TestRenderStripsDuplicatedMarkdownHeading(t *testing.T) {
	driver := New(nil, nil)

	rendered, err := driver.Render(channel.RenderInput{
		ContentTitle: "Weekly Update",
		ContentBody:  "# Weekly Update\n\nHello team",
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if rendered.Title != "Weekly Update" {
		t.Fatalf("Title = %q", rendered.Title)
	}
	if rendered.Body != "Hello team" {
		t.Fatalf("Body = %q", rendered.Body)
	}
	if rendered.RenderMode != domain.RenderModeFeishuPost {
		t.Fatalf("RenderMode = %q", rendered.RenderMode)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
