package email

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/dsmes/dsmes-backend/templates"
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
	// 1. Try embedded templates first (self-contained in binary)
	tmpl, err := template.ParseFS(templates.FS, name)
	if err == nil {
		var buf bytes.Buffer
		if execErr := tmpl.ExecuteTemplate(&buf, name, data); execErr == nil {
			return buf.String(), nil
		}
	}

	// 2. Fallback to filesystem if embedded lookup failed
	path := filepath.Join(t.basePath, name)
	tmpl, err = template.ParseFiles(path)
	if err != nil {
		return "", fmt.Errorf("email: failed to parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("email: failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}
