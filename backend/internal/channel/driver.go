package channel

import "context"

type Driver interface {
	Type() string
	ValidateAccount(config map[string]any, secretRef string) error
	ValidateTarget(config map[string]any, targetKey string) error
	Render(input RenderInput) (RenderedMessage, error)
	Send(ctx context.Context, request SendRequest) (SendResult, error)
}

type RenderInput struct {
	TemplateName string
	ContentTitle string
	ContentBody  string
	Frontmatter  map[string]any
	TargetName   string
	TargetKey    string
}

type RenderedMessage struct {
	Title      string
	Body       string
	RenderMode string
}

type SendRequest struct {
	SecretValue string
	TargetKey   string
	Body        string
	Config      map[string]any
}

type SendResult struct {
	ExternalMessageID string
	ProviderResponse  string
}
