package render

import (
	"strings"
	"testing"
)

func TestRenderTemplateSupportsYAMLArrayTags(t *testing.T) {
	renderer := NewTemplateRenderer()

	var context Context
	context.Content.Title = "Hello"
	context.Content.BodyMarkdown = "Body"
	context.Meta = map[string]any{
		"tags": []any{"go", "telegram"},
	}

	result, err := renderer.RenderTemplate("default", `{{ .Content.BodyMarkdown }}

{{ with .Meta.tags }}{{ hashtags . }}{{ end }}`, context)
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}

	if !strings.Contains(result, "#go #telegram") {
		t.Fatalf("RenderTemplate() = %q, want tags output", result)
	}
}
