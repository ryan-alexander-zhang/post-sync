package feishu

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/erpang/post-sync/internal/channel"
	"github.com/erpang/post-sync/internal/domain"
)

const defaultBaseURL = "https://open.feishu.cn"

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
		return fmt.Errorf("feishu enterprise secretRef is required")
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
	return renderPostMessage(input), nil
}

func (d *Driver) Send(ctx context.Context, request channel.SendRequest) (channel.SendResult, error) {
	token, err := d.tokenProvider.GetTenantAccessToken(ctx, request.Account)
	if err != nil {
		return channel.SendResult{}, err
	}

	baseURL := strings.TrimSpace(stringValue(request.Account.Config["baseUrl"]))
	if baseURL == "" {
		baseURL = defaultBaseURL
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
	if _, err := rawBody.ReadFrom(response.Body); err != nil {
		return channel.SendResult{}, err
	}

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

type PersonalDriver struct {
	client *http.Client
}

func NewPersonal(client *http.Client) *PersonalDriver {
	if client == nil {
		client = http.DefaultClient
	}

	return &PersonalDriver{client: client}
}

func (d *PersonalDriver) Type() string {
	return domain.ChannelTypePersonalFeishu
}

func (d *PersonalDriver) ValidateAccount(input channel.AccountValidationInput) error {
	webhookURL, err := resolveConfiguredValueAllowDirect(input.SecretRef, input.Config, "webhookUrl")
	if err != nil {
		return err
	}
	if strings.TrimSpace(webhookURL) == "" {
		return fmt.Errorf("personal feishu webhook url is required")
	}

	if _, err := resolveConfiguredValue(input.Config, "signSecretRef", "signSecret"); err != nil {
		return err
	}
	return nil
}

func (d *PersonalDriver) NormalizeTarget(input channel.TargetInput) (channel.NormalizedTarget, error) {
	webhookRef := strings.TrimSpace(stringValue(input.Config["webhookEnvRef"]))
	if webhookRef == "" {
		webhookRef = strings.TrimSpace(stringValue(input.Config["webhookUrl"]))
	}
	if webhookRef == "" {
		webhookRef = strings.TrimSpace(input.TargetKey)
	}
	if webhookRef == "" {
		return channel.NormalizedTarget{}, fmt.Errorf("personal feishu webhook reference is required")
	}

	targetType := strings.TrimSpace(input.TargetType)
	if targetType == "" {
		targetType = domain.TargetTypePersonalFeishuWebhook
	}
	if targetType != domain.TargetTypePersonalFeishuWebhook {
		return channel.NormalizedTarget{}, fmt.Errorf("unsupported personal feishu target type")
	}

	return channel.NormalizedTarget{
		TargetType: targetType,
		TargetKey:  buildWebhookTargetKey(webhookRef),
		Config: map[string]any{
			"webhookEnvRef": webhookRef,
		},
	}, nil
}

func (d *PersonalDriver) Render(input channel.RenderInput) (channel.RenderedMessage, error) {
	body := stripDuplicatedTitle(input.ContentTitle, input.ContentBody)
	text := strings.TrimSpace(body)
	title := strings.TrimSpace(input.ContentTitle)
	if title != "" && text != "" {
		text = title + "\n\n" + text
	} else if title != "" {
		text = title
	}

	return channel.RenderedMessage{
		Title:      title,
		Body:       text,
		RenderMode: domain.RenderModeFeishuText,
	}, nil
}

func (d *PersonalDriver) Send(ctx context.Context, request channel.SendRequest) (channel.SendResult, error) {
	webhookURL, err := resolveConfiguredValueAllowDirect(
		request.Account.SecretRef,
		request.Account.Config,
		"webhookUrl",
	)
	if err != nil {
		return channel.SendResult{}, err
	}

	timestamp := time.Now().Unix()
	signSecret, err := resolveConfiguredValue(request.Account.Config, "signSecretRef", "signSecret")
	if err != nil {
		return channel.SendResult{}, err
	}

	sign, err := genWebhookSign(signSecret, timestamp)
	if err != nil {
		return channel.SendResult{}, err
	}

	payloadBody, err := json.Marshal(map[string]any{
		"timestamp": fmt.Sprintf("%d", timestamp),
		"sign":      sign,
		"msg_type":  "text",
		"content": map[string]string{
			"text": request.Body,
		},
	})
	if err != nil {
		return channel.SendResult{}, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payloadBody))
	if err != nil {
		return channel.SendResult{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := d.client.Do(httpRequest)
	if err != nil {
		return channel.SendResult{}, err
	}
	defer response.Body.Close()

	var rawBody bytes.Buffer
	if _, err := rawBody.ReadFrom(response.Body); err != nil {
		return channel.SendResult{}, err
	}

	var payloadResponse struct {
		Code    int    `json:"code"`
		Message string `json:"msg"`
		Status  string `json:"StatusMessage"`
	}
	if err := json.Unmarshal(rawBody.Bytes(), &payloadResponse); err != nil {
		return channel.SendResult{}, err
	}

	if response.StatusCode >= 400 || payloadResponse.Code != 0 {
		message := strings.TrimSpace(payloadResponse.Message)
		if message == "" {
			message = strings.TrimSpace(payloadResponse.Status)
		}
		if message == "" {
			message = "unknown error"
		}
		return channel.SendResult{}, fmt.Errorf("personal feishu send failed: code=%d msg=%s", payloadResponse.Code, message)
	}

	return channel.SendResult{
		ExternalMessageID: request.Target.Key,
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

func renderPostMessage(input channel.RenderInput) channel.RenderedMessage {
	return channel.RenderedMessage{
		Title:      input.ContentTitle,
		Body:       stripDuplicatedTitle(input.ContentTitle, input.ContentBody),
		RenderMode: domain.RenderModeFeishuPost,
	}
}

func buildWebhookTargetKey(webhookURL string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(webhookURL)))
	return "webhook:" + hex.EncodeToString(sum[:8])
}

func genWebhookSign(secret string, timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	if _, err := h.Write(nil); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func resolveConfiguredValueAllowDirect(secretRef string, config map[string]any, directKey string) (string, error) {
	if trimmed := strings.TrimSpace(secretRef); trimmed != "" {
		if value := strings.TrimSpace(os.Getenv(trimmed)); value != "" {
			return value, nil
		}
		return "", fmt.Errorf("environment variable %s is not set", trimmed)
	}
	if value := strings.TrimSpace(stringValue(config[directKey])); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("%s is required", directKey)
}
