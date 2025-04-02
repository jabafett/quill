package factories

import (
        "bytes"
        "fmt"
        "text/template"

        "github.com/jabafett/quill/internal/utils/templates"
)

// TemplateType identifies different types of templates
type TemplateType string

const (
        CommitMessageType TemplateType = "CommitMessage"
        SuggestionType    TemplateType = "Suggestion"
)

// TemplateFactory manages template creation and rendering
type TemplateFactory struct {
        templates map[TemplateType]*template.Template
}

// NewTemplateFactory creates a new template factory instance
func NewTemplateFactory() (*TemplateFactory, error) {
        factory := &TemplateFactory{
                templates: make(map[TemplateType]*template.Template),
        }

        // Initialize templates
        if err := factory.initializeTemplates(); err != nil {
                return nil, fmt.Errorf("failed to initialize templates: %w", err)
        }

        return factory, nil
}

// initializeTemplates loads all templates into memory
func (f *TemplateFactory) initializeTemplates() error {
        templateMap := map[TemplateType]string{
                CommitMessageType: templates.CommitMessageTemplate,
                SuggestionType:    templates.SuggestTemplate,
        }

        for typ, content := range templateMap {
                tmpl, err := template.New(string(typ)).Option("missingkey=error").Parse(content)
                if err != nil {
                        return fmt.Errorf("failed to parse template %s: %w", typ, err)
                }
                f.templates[typ] = tmpl
        }
        return nil
}

// ValidateTemplates ensures all templates are valid
func ValidateTemplates() error {
        templates := map[string]string{
                "CommitMessage": templates.CommitMessageTemplate,
                "Suggest":       templates.SuggestTemplate,
        }

        for name, content := range templates {
                if _, err := template.New(name).Parse(content); err != nil {
                        return err
                }
        }

        return nil
}

// Generate creates a filled template of the specified type
func (f *TemplateFactory) Generate(typ TemplateType, data interface{}) (string, error) {
        tmpl, exists := f.templates[typ]
        if !exists {
                return "", fmt.Errorf("template type %s not found", typ)
        }

        var buf bytes.Buffer
        if err := tmpl.Execute(&buf, data); err != nil {
                return "", fmt.Errorf("failed to execute template: %w", err)
        }

        return buf.String(), nil
}
