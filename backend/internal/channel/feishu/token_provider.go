package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/erpang/post-sync/internal/channel"
)

type TokenProvider struct {
	client *http.Client
	mu     sync.Mutex
	cache  map[string]cachedToken
}

type cachedToken struct {
	value     string
	expiresAt time.Time
}

func NewTokenProvider(client *http.Client) *TokenProvider {
	if client == nil {
		client = http.DefaultClient
	}

	return &TokenProvider{
		client: client,
		cache:  map[string]cachedToken{},
	}
}

func (p *TokenProvider) GetTenantAccessToken(ctx context.Context, account channel.Account) (string, error) {
	if tokenEnv := strings.TrimSpace(stringValue(account.Config["tokenEnv"])); tokenEnv != "" {
		if token := strings.TrimSpace(os.Getenv(tokenEnv)); token != "" {
			return token, nil
		}
	}

	appID, err := resolveConfiguredValue(account.Config, "appIdEnv", "appId")
	if err != nil {
		return "", err
	}

	appSecret := strings.TrimSpace(os.Getenv(account.SecretRef))
	if appSecret == "" {
		return "", fmt.Errorf("feishu app secret env %s is empty", account.SecretRef)
	}

	baseURL := strings.TrimSpace(stringValue(account.Config["baseUrl"]))
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	cacheKey := fmt.Sprintf("%s|%s", strings.TrimRight(baseURL, "/"), appID)
	if token, ok := p.getCachedToken(cacheKey); ok {
		return token, nil
	}

	requestBody, err := json.Marshal(map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
	})
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/open-apis/auth/v3/tenant_access_token/internal"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := p.client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var payload struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}

	if response.StatusCode >= 400 || payload.Code != 0 {
		return "", fmt.Errorf("feishu token fetch failed: code=%d msg=%s", payload.Code, payload.Msg)
	}
	if strings.TrimSpace(payload.TenantAccessToken) == "" {
		return "", fmt.Errorf("feishu token fetch failed: empty tenant_access_token")
	}

	expiresAt := time.Now().Add(time.Duration(payload.Expire) * time.Second)
	p.storeCachedToken(cacheKey, cachedToken{
		value:     payload.TenantAccessToken,
		expiresAt: expiresAt,
	})

	return payload.TenantAccessToken, nil
}

func (p *TokenProvider) getCachedToken(key string) (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	token, ok := p.cache[key]
	if !ok {
		return "", false
	}
	if time.Until(token.expiresAt) <= 30*time.Minute {
		delete(p.cache, key)
		return "", false
	}
	return token.value, true
}

func (p *TokenProvider) storeCachedToken(key string, token cachedToken) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cache[key] = token
}

func resolveConfiguredValue(config map[string]any, envKey string, directKey string) (string, error) {
	if envName := strings.TrimSpace(stringValue(config[envKey])); envName != "" {
		if value := strings.TrimSpace(os.Getenv(envName)); value != "" {
			return value, nil
		}
		return "", fmt.Errorf("environment variable %s is not set", envName)
	}
	if value := strings.TrimSpace(stringValue(config[directKey])); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("%s is required", envKey)
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}
