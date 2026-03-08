package render

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"strings"
	"text/template"

	"github.com/yuin/goldmark"
)

var (
	paragraphOpen  = regexp.MustCompile(`<p>`)
	paragraphClose = regexp.MustCompile(`</p>`)
	breakLine      = regexp.MustCompile(`<br\s*/?>`)
)

type Context struct {
	Content struct {
		ID           string
		Title        string
		BodyMarkdown string
		BodyPlain    string
	}
	Meta    map[string]any
	Channel struct {
		TargetName string
		TargetKey  string
	}
}

type TemplateRenderer struct {
	engine goldmark.Markdown
}

func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{engine: goldmark.New()}
}

func (r *TemplateRenderer) RenderMarkdown(templateName, text string, context Context) (string, error) {
	tpl, err := template.New(templateName).Funcs(template.FuncMap{
		"join": joinValues,
	}).Parse(text)
	if err != nil {
		return "", err
	}

	var markdownBuffer bytes.Buffer
	if err := tpl.Execute(&markdownBuffer, context); err != nil {
		return "", err
	}

	var htmlBuffer bytes.Buffer
	if err := r.engine.Convert(markdownBuffer.Bytes(), &htmlBuffer); err != nil {
		return "", err
	}

	result := htmlBuffer.String()
	result = paragraphOpen.ReplaceAllString(result, "")
	result = paragraphClose.ReplaceAllString(result, "\n\n")
	result = breakLine.ReplaceAllString(result, "\n")
	result = strings.TrimSpace(result)
	return sanitizeTelegramHTML(result), nil
}

func sanitizeTelegramHTML(input string) string {
	replacer := strings.NewReplacer(
		"<h1>", "<b>", "</h1>", "</b>\n\n",
		"<h2>", "<b>", "</h2>", "</b>\n\n",
		"<h3>", "<b>", "</h3>", "</b>\n\n",
		"<ul>", "", "</ul>", "",
		"<ol>", "", "</ol>", "",
		"<li>", "• ", "</li>", "\n",
		"<strong>", "<b>", "</strong>", "</b>",
		"<em>", "<i>", "</em>", "</i>",
		"&lt;", "<", "&gt;", ">",
	)

	output := replacer.Replace(input)
	output = strings.ReplaceAll(output, "<pre><code>", "<pre>")
	output = strings.ReplaceAll(output, "</code></pre>", "</pre>")
	return strings.TrimSpace(output)
}

func EscapeFallback(text string) string {
	return strings.ReplaceAll(html.EscapeString(text), "\n", "\n")
}

func joinValues(value any, separator string) string {
	switch typed := value.(type) {
	case []string:
		return strings.Join(typed, separator)
	case []any:
		items := make([]string, 0, len(typed))
		for _, entry := range typed {
			text := strings.TrimSpace(fmt.Sprint(entry))
			if text == "" {
				continue
			}
			items = append(items, text)
		}
		return strings.Join(items, separator)
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}
