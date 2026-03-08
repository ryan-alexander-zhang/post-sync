package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/erpang/post-sync/internal/channel"
	"github.com/erpang/post-sync/internal/domain"
)

type Driver struct {
	client        *http.Client
	tokenProvider *TokenProvider
}

func New(client *http.Client, tokenProvider *TokenProvider) *Driver {
	if client == nil {
		client = http.DefaultClient
	}
	if tokenProvider == nil {
		tokenProvider = NewTokenProvider(client)
	}

	return &Driver{
		client:        client,
		tokenProvider: tokenProvider,
	}
}

func (d *Driver) Type() string {
	return domain.ChannelTypeFeishu
}

func (d *Driver) ValidateAccount(input channel.AccountValidationInput) error {
	secretRef := strings.TrimSpace(input.SecretRef)
	if secretRef == "" {
		return fmt.Errorf("feishu secretRef is required")
	}

	if tokenEnv := strings.TrimSpace(stringValue(input.Config["tokenEnv"])); tokenEnv != "" {
		if value := strings.TrimSpace(os.Getenv(tokenEnv)); value != "" {
			return nil
		}
	}

	if _, err := resolveConfiguredValue(input.Config, "appIdEnv", "appId"); err != nil {
		return err
	}
	if _, ok := os.LookupEnv(secretRef); !ok {
		return fmt.Errorf("environment variable %s is not set", secretRef)
	}
	return nil
}

func (d *Driver) NormalizeTarget(input channel.TargetInput) (channel.NormalizedTarget, error) {
	chatID := strings.TrimSpace(input.TargetKey)
	if value := strings.TrimSpace(stringValue(input.Config["chatId"])); value != "" {
		chatID = value
	}
	if chatID == "" {
		return channel.NormalizedTarget{}, fmt.Errorf("feishu chat id is required")
	}

	targetType := strings.TrimSpace(input.TargetType)
	if targetType == "" {
		targetType = domain.TargetTypeFeishuChat
	}
	if targetType != domain.TargetTypeFeishuChat {
		return channel.NormalizedTarget{}, fmt.Errorf("unsupported feishu target type")
	}

	return channel.NormalizedTarget{
		TargetType: targetType,
		TargetKey:  chatID,
		Config: map[string]any{
			"receiveIdType": "chat_id",
			"chatId":        chatID,
		},
	}, nil
}

func (d *Driver) Render(input channel.RenderInput) (channel.RenderedMessage, error) {
	return channel.RenderedMessage{
		Title:      input.ContentTitle,
		Body:       stripDuplicatedTitle(input.ContentTitle, input.ContentBody),
		RenderMode: domain.RenderModeFeishuPost,
	}, nil
}

func (d *Driver) Send(ctx context.Context, request channel.SendRequest) (channel.SendResult, error) {
	token, err := d.tokenProvider.GetTenantAccessToken(ctx, request.Account)
	if err != nil {
		return channel.SendResult{}, err
	}

	baseURL := strings.TrimSpace(stringValue(request.Account.Config["baseUrl"]))
	if baseURL == "" {
		baseURL = "https://open.feishu.cn"
	}

	receiveID := request.Target.Key
	if value := strings.TrimSpace(stringValue(request.Target.Config["chatId"])); value != "" {
		receiveID = value
	}
	if receiveID == "" {
		return channel.SendResult{}, fmt.Errorf("feishu chat id is required")
	}

	contentBody, err := json.Marshal(buildPostContent(request.Title, request.Body))
	if err != nil {
		return channel.SendResult{}, err
	}

	payload := map[string]any{
		"receive_id": receiveID,
		"msg_type":   "post",
		"content":    string(contentBody),
	}
	if request.IdempotencyKey != "" {
		payload["uuid"] = truncate(request.IdempotencyKey, 50)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return channel.SendResult{}, err
	}

	receiveIDType := strings.TrimSpace(stringValue(request.Target.Config["receiveIdType"]))
	if receiveIDType == "" {
		receiveIDType = "chat_id"
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/open-apis/im/v1/messages?receive_id_type=" + url.QueryEscape(receiveIDType)
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return channel.SendResult{}, err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+token)
	httpRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := d.client.Do(httpRequest)
	if err != nil {
		return channel.SendResult{}, err
	}
	defer response.Body.Close()

	var rawBody bytes.Buffer
	reader := bytes.NewBuffer(nil)
	if _, err := reader.ReadFrom(response.Body); err != nil {
		return channel.SendResult{}, err
	}
	rawBody = *reader

	var payloadResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			MessageID string `json:"message_id"`
		} `json:"data"`
		Error struct {
			LogID string `json:"log_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rawBody.Bytes(), &payloadResponse); err != nil {
		return channel.SendResult{}, err
	}

	if response.StatusCode >= 400 || payloadResponse.Code != 0 {
		logID := strings.TrimSpace(payloadResponse.Error.LogID)
		if logID != "" {
			return channel.SendResult{}, fmt.Errorf("feishu send failed: code=%d msg=%s log_id=%s", payloadResponse.Code, payloadResponse.Msg, logID)
		}
		return channel.SendResult{}, fmt.Errorf("feishu send failed: code=%d msg=%s", payloadResponse.Code, payloadResponse.Msg)
	}

	return channel.SendResult{
		ExternalMessageID: payloadResponse.Data.MessageID,
		ProviderResponse:  rawBody.String(),
	}, nil
}

func truncate(input string, limit int) string {
	if len(input) <= limit {
		return input
	}
	return input[:limit]
}

func buildPostContent(title, body string) map[string]any {
	content := map[string]any{
		"zh_cn": map[string]any{
			"content": [][]map[string]string{
				{
					{
						"tag":  "md",
						"text": strings.TrimSpace(body),
					},
				},
			},
		},
	}

	if trimmedTitle := strings.TrimSpace(title); trimmedTitle != "" {
		content["zh_cn"].(map[string]any)["title"] = trimmedTitle
	}

	return content
}

func stripDuplicatedTitle(title, body string) string {
	trimmedBody := strings.TrimSpace(body)
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" || trimmedBody == "" {
		return trimmedBody
	}

	prefix := "# " + trimmedTitle
	if trimmedBody == prefix {
		return ""
	}
	if strings.HasPrefix(trimmedBody, prefix+"\n") {
		return strings.TrimSpace(strings.TrimPrefix(trimmedBody, prefix))
	}

	return trimmedBody
}
