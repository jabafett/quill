package factories

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/jabafett/quill/internal/utils/templates"
)

// Additional template types
const (
	ContextExtractionType TemplateType = "ContextExtraction"
	AnalysisType          TemplateType = "Analysis"
	SuggestionType        TemplateType = "Suggestion"
)

// Additional template data types
type (
	// ContextData holds data for context extraction templates
	ContextData struct {
		Files       []string
		FileContent map[string]string
		GitHistory  string
	}

	// AnalysisData holds data for analysis templates
	AnalysisData struct {
		Context    string
		Changes    string
		FileTypes  map[string]string
		Complexity int
	}

	// SuggestionData holds data for suggestion templates
	SuggestionData struct {
		Context     string
		Analysis    string
		Constraints []string
	}
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
		CommitMessageType:     templates.CommitMessageTemplate,
		ContextExtractionType: templates.ContextExtractionTemplate,
		SuggestionType:        templates.SuggestTemplate,
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
		"ContextExtraction": templates.ContextExtractionTemplate,
		"CommitMessage":     templates.CommitMessageTemplate,
		"Suggest":           templates.SuggestTemplate,
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

// GenerateContext creates a context extraction prompt
func (f *TemplateFactory) GenerateContext(data ContextData) (string, error) {
	return f.generateFromType(ContextExtractionType, data)
}

// GenerateAnalysis creates an analysis prompt
func (f *TemplateFactory) GenerateAnalysis(data AnalysisData) (string, error) {
	return f.generateFromType(AnalysisType, data)
}

// GenerateSuggestion creates a suggestion prompt
func (f *TemplateFactory) GenerateSuggestion(data SuggestionData) (string, error) {
	return f.generateFromType(SuggestionType, data)
}

// generateFromType is a generic method to generate prompts
func (f *TemplateFactory) generateFromType(typ TemplateType, data interface{}) (string, error) {
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
