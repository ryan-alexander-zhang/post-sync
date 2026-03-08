package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	frontmatterPattern = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n?`)
	trailingSpace      = regexp.MustCompile(`[ \t]+\n`)
	multiBlankLines    = regexp.MustCompile(`\n{3,}`)
	markdownNoise      = regexp.MustCompile(`(?m)^\s{0,3}(#{1,6}|\*|-|\d+\.)\s*`)
	linkPattern        = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
)

type ParsedMarkdown struct {
	Title           string
	Frontmatter     map[string]any
	FrontmatterJSON string
	BodyMarkdown    string
	BodyPlain       string
	BodyHash        string
}

func ParseMarkdown(raw []byte) (ParsedMarkdown, error) {
	text := normalizeNewlines(string(raw))
	frontmatter, body, err := splitFrontmatter(text)
	if err != nil {
		return ParsedMarkdown{}, err
	}

	body = normalizeBody(body)
	frontmatterJSON, err := marshalFrontmatter(frontmatter)
	if err != nil {
		return ParsedMarkdown{}, err
	}

	title := firstString(frontmatter["title"])
	if title == "" {
		title = inferTitle(body)
	}

	return ParsedMarkdown{
		Title:           title,
		Frontmatter:     frontmatter,
		FrontmatterJSON: frontmatterJSON,
		BodyMarkdown:    body,
		BodyPlain:       toPlainText(body),
		BodyHash:        hash(body),
	}, nil
}

func splitFrontmatter(text string) (map[string]any, string, error) {
	matches := frontmatterPattern.FindStringSubmatch(text)
	if len(matches) == 0 {
		return map[string]any{}, text, nil
	}

	frontmatter := map[string]any{}
	if err := yaml.Unmarshal([]byte(matches[1]), &frontmatter); err != nil {
		return nil, "", err
	}

	return frontmatter, text[len(matches[0]):], nil
}

func marshalFrontmatter(frontmatter map[string]any) (string, error) {
	if len(frontmatter) == 0 {
		return "{}", nil
	}

	data, err := json.Marshal(frontmatter)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func normalizeNewlines(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	return strings.ReplaceAll(input, "\r", "\n")
}

func normalizeBody(input string) string {
	trimmed := strings.TrimSpace(input)
	trimmed = trailingSpace.ReplaceAllString(trimmed, "\n")
	trimmed = multiBlankLines.ReplaceAllString(trimmed, "\n\n")
	return trimmed
}

func toPlainText(input string) string {
	replacedLinks := linkPattern.ReplaceAllString(input, "$1")
	cleaned := markdownNoise.ReplaceAllString(replacedLinks, "")
	cleaned = strings.ReplaceAll(cleaned, "`", "")
	cleaned = strings.ReplaceAll(cleaned, "*", "")
	cleaned = strings.ReplaceAll(cleaned, "_", "")
	return strings.TrimSpace(cleaned)
}

func hash(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

func inferTitle(body string) string {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		candidate := strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

func firstString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}
