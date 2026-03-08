package telegram

import (
	"context"
	"fmt"
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
	body := strings.TrimSpace(input.ContentBody)
	if strings.TrimSpace(input.ContentTitle) != "" {
		body = "<b>" + input.ContentTitle + "</b>\n\n" + body
	}

	return channel.RenderedMessage{
		Title:      input.ContentTitle,
		Body:       body,
		RenderMode: domain.RenderModeTelegram,
	}, nil
}

func (d *Driver) Send(_ context.Context, request channel.SendRequest) (channel.SendResult, error) {
	return channel.SendResult{
		ExternalMessageID: "",
		ProviderResponse:  `{"mode":"stub"}`,
	}, nil
}
