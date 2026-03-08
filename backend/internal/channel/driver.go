package channel

import "context"

type Driver interface {
	Type() string
	ValidateAccount(input AccountValidationInput) error
	NormalizeTarget(input TargetInput) (NormalizedTarget, error)
	Render(input RenderInput) (RenderedMessage, error)
	Send(ctx context.Context, request SendRequest) (SendResult, error)
}

type AccountValidationInput struct {
	SecretRef string
	Config    map[string]any
}

type TargetInput struct {
	TargetType string
	TargetKey  string
	TargetName string
	Config     map[string]any
}

type NormalizedTarget struct {
	TargetType string
	TargetKey  string
	Config     map[string]any
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

type Account struct {
	ID        string
	Type      string
	Name      string
	SecretRef string
	Config    map[string]any
	IsEnabled bool
}

type Target struct {
	ID        string
	Type      string
	Key       string
	Name      string
	Config    map[string]any
	IsEnabled bool
}

type SendRequest struct {
	Account        Account
	Target         Target
	Body           string
	IdempotencyKey string
}

type SendResult struct {
	ExternalMessageID string
	ProviderResponse  string
}
