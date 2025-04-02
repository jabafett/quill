package context

import (
	"regexp"
	"strings"
)

// Common patterns for import statements across languages
var (
	// Remove common prefixes/suffixes
	importPrefixes = []string{
		"import ",
		"use ",
		"#include ",
		"from ",
		"require ",
	}

	importSuffixes = []string{
		";",
		"\n",
	}

	// Clean up patterns
	cleanPatterns = []*regexp.Regexp{
		regexp.MustCompile(`^["'<]`),        // Leading quotes/brackets
		regexp.MustCompile(`["'>]$`),        // Trailing quotes/brackets
		regexp.MustCompile(`\s+as\s+.*$`),   // Remove 'as' aliases
		regexp.MustCompile(`\s+from\s+.*$`), // Keep only the 'from' part
		regexp.MustCompile(`^_\s*`),         // Remove Go's underscore imports
	}
)

// normalizeImport cleans and normalizes import strings across different languages
func normalizeImport(imp string) string {
	// Trim all whitespace
	imp = strings.TrimSpace(imp)
	if imp == "" {
		return ""
	}

	// Remove common prefixes
	for _, prefix := range importPrefixes {
		if strings.HasPrefix(imp, prefix) {
			imp = strings.TrimSpace(strings.TrimPrefix(imp, prefix))
		}
	}

	// Remove common suffixes
	for _, suffix := range importSuffixes {
		if strings.HasSuffix(imp, suffix) {
			imp = strings.TrimSpace(strings.TrimSuffix(imp, suffix))
		}
	}

	// Apply cleanup patterns
	for _, pattern := range cleanPatterns {
		imp = strings.TrimSpace(pattern.ReplaceAllString(imp, ""))
	}

	return imp
}

// ImportSet manages a unique set of imports
type ImportSet struct {
	imports map[string]bool // Set of normalized imports
}

// NewImportSet creates a new import set
func NewImportSet() *ImportSet {
	return &ImportSet{
		imports: make(map[string]bool),
	}
}

// Add adds an import to the set and returns true if it was not already present
func (s *ImportSet) Add(imp string) bool {
	if imp == "" {
		return false
	}

	// Normalize the import
	norm := normalizeImport(imp)
	if norm == "" {
		return false
	}

	// Check if we've seen this normalized form
	if s.imports[norm] {
		return false
	}

	// Add the normalized import
	s.imports[norm] = true
	return true
}

// Imports returns the list of normalized imports
func (s *ImportSet) Imports() []string {
	imports := make([]string, 0, len(s.imports))
	for imp := range s.imports {
		imports = append(imports, imp)
	}
	return imports
}

// Dependencies returns the list of unique dependency names
func (s *ImportSet) Dependencies() []string {
	return s.Imports()
}
