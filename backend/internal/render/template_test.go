package render

import (
	"strings"
	"testing"
)

func TestRenderMarkdownSupportsYAMLArrayTags(t *testing.T) {
	renderer := NewTemplateRenderer()

	var context Context
	context.Content.Title = "Hello"
	context.Content.BodyMarkdown = "Body"
	context.Meta = map[string]any{
		"tags": []any{"go", "telegram"},
	}

	result, err := renderer.RenderMarkdown("default", `{{ .Content.BodyMarkdown }}

{{ with .Meta.tags }}{{ hashtags . }}{{ end }}`, context)
	if err != nil {
		t.Fatalf("RenderMarkdown() error = %v", err)
	}

	if !strings.Contains(result, "#go #telegram") {
		t.Fatalf("RenderMarkdown() = %q, want tags output", result)
	}
}
