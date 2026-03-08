package parser

import "testing"

func TestParseMarkdownDropsFrontmatterAndNormalizesBody(t *testing.T) {
	input := []byte("---\ntitle: Sample\ntags:\n  - a\n---\nHello\r\n\r\n\r\nWorld   \n")

	parsed, err := ParseMarkdown(input)
	if err != nil {
		t.Fatalf("ParseMarkdown() error = %v", err)
	}

	if parsed.Title != "Sample" {
		t.Fatalf("Title = %q, want Sample", parsed.Title)
	}

	if parsed.BodyMarkdown != "Hello\n\nWorld" {
		t.Fatalf("BodyMarkdown = %q", parsed.BodyMarkdown)
	}

	if parsed.BodyHash == "" {
		t.Fatal("BodyHash should not be empty")
	}
}

func TestParseMarkdownDoesNotInferTitleFromBody(t *testing.T) {
	input := []byte("# Body Heading\n\nContent")

	parsed, err := ParseMarkdown(input)
	if err != nil {
		t.Fatalf("ParseMarkdown() error = %v", err)
	}

	if parsed.Title != "" {
		t.Fatalf("Title = %q, want empty", parsed.Title)
	}
}
