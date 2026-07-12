package email

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
)

// TemplateEngine handles loading and rendering HTML templates.
type TemplateEngine struct {
	basePath string
}

// NewTemplateEngine constructs a TemplateEngine.
func NewTemplateEngine(basePath string) *TemplateEngine {
	return &TemplateEngine{basePath: basePath}
}

// Render renders the template at name using the provided data.
func (t *TemplateEngine) Render(name string, data any) (string, error) {
	path := filepath.Join(t.basePath, name)
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", fmt.Errorf("email: failed to parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("email: failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}
