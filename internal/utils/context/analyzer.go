package context

import (
	"context"
	"fmt"
	"os"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/lua"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
)

// Analyzer interface for language-specific analyzers
type Analyzer interface {
	Analyze(path string) (*FileContext, error)
}

// DefaultAnalyzer provides tree-sitter based file analysis
type DefaultAnalyzer struct {
	parsers      map[string]*sitter.Parser
	languages    map[string]*sitter.Language
	typeDetector *FileTypeDetector
	langLoaders  map[string]func() *sitter.Language
}

// NewDefaultAnalyzer creates a new analyzer with initialized parsers
func NewDefaultAnalyzer() *DefaultAnalyzer {
	a := &DefaultAnalyzer{
		parsers:      make(map[string]*sitter.Parser),
		languages:    make(map[string]*sitter.Language),
		typeDetector: &FileTypeDetector{},
		langLoaders:  make(map[string]func() *sitter.Language),
	}

	// Register language loaders instead of initializing them
	a.registerLanguages()
	return a
}

func (a *DefaultAnalyzer) registerLanguages() {
	a.langLoaders = map[string]func() *sitter.Language{
		"go":         golang.GetLanguage,
		"javascript": javascript.GetLanguage,
		"typescript": javascript.GetLanguage,
		"python":     python.GetLanguage,
		"cpp":        cpp.GetLanguage,
		"c++":        cpp.GetLanguage,
		"rust":       rust.GetLanguage,
		"java":       java.GetLanguage,
		"ruby":       ruby.GetLanguage,
		"css":        css.GetLanguage,
		"html":       html.GetLanguage,
		"lua":        lua.GetLanguage,
	}
}

func (a *DefaultAnalyzer) getLanguage(fileType string) (*sitter.Language, error) {
	// Check if language is already initialized
	if lang, ok := a.languages[fileType]; ok {
		return lang, nil
	}

	// Check if we have a loader for this language
	loader, ok := a.langLoaders[fileType]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", fileType)
	}

	// Initialize the language
	lang := loader()
	a.languages[fileType] = lang
	
	// Initialize the parser
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	a.parsers[fileType] = parser

	return lang, nil
}

func (a *DefaultAnalyzer) Analyze(path string) (*FileContext, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	fileType := a.typeDetector.DetectFileType(path, content)
	
	// Get or initialize the language
	lang, err := a.getLanguage(fileType)
	if err != nil {
		return a.basicAnalyze(path, content, fileType)
	}

	parser := a.parsers[fileType]
	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}
	defer tree.Close()

	symbols := a.extractSymbolsFromTree(tree.RootNode(), content, fileType, lang)
	imports := a.findImports(tree.RootNode(), content, fileType, lang)
	complexity := a.calculateNodeComplexity(tree.RootNode())

	return &FileContext{
		Path:       path,
		Type:       fileType,
		Symbols:    symbols,
		Imports:    imports,
		Complexity: complexity,
		UpdatedAt:  time.Now(),
		AST:        tree.RootNode().String(),
	}, nil
}

func (a *DefaultAnalyzer) extractSymbolsFromTree(node *sitter.Node, content []byte, fileType string, lang *sitter.Language) []Symbol {
	var symbols []Symbol
	queryStr := a.getSymbolQueryForLanguage(fileType)
	if queryStr == "" {
		return symbols
	}

	query, err := sitter.NewQuery([]byte(queryStr), lang)
	if err != nil {
		return symbols
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, node)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		match = qc.FilterPredicates(match, content)
		for _, capture := range match.Captures {
			symbol := Symbol{
				Name:      string(capture.Node.Content(content)),
				Type:      capture.Node.Type(),
				StartLine: int(capture.Node.StartPoint().Row) + 1,
				EndLine:   int(capture.Node.EndPoint().Row) + 1,
			}
			symbol.Complexity = a.calculateNodeComplexity(capture.Node)
			symbols = append(symbols, symbol)
		}
	}

	return symbols
}

func (a *DefaultAnalyzer) getSymbolQueryForLanguage(fileType string) string {
	queries := map[string]string{
		"go": `
			(function_declaration
				name: (identifier) @func.name) @function
			(method_declaration
				name: (field_identifier) @method.name) @method
			(type_declaration 
				(type_spec 
					name: (type_identifier) @type.name
					type: [(struct_type) (interface_type)] @type.kind)) @type
		`,
		"javascript": `
			(function_declaration
				name: (identifier) @func.name) @function
			(class_declaration
				name: (identifier) @class.name) @class
			(method_definition
				name: (property_identifier) @method.name) @method
			(arrow_function
				name: (identifier) @arrow.name) @arrow
		`,
		"python": `
			(function_definition
				name: (identifier) @func.name) @function
			(class_definition
				name: (identifier) @class.name) @class
			(decorated_definition) @decorated
		`,
		"rust": `
			(function_item
				name: (identifier) @func.name) @function
			(struct_item
				name: (type_identifier) @struct.name) @struct
			(impl_item) @impl
			(trait_item
				name: (type_identifier) @trait.name) @trait
		`,
		"java": `
			(method_declaration
				name: (identifier) @method.name) @method
			(class_declaration
				name: (identifier) @class.name) @class
			(interface_declaration
				name: (identifier) @interface.name) @interface
		`,
		"ruby": `
			(method
				name: (identifier) @method.name) @method
			(class
				name: (constant) @class.name) @class
			(module
				name: (constant) @module.name) @module
		`,
		"cpp": `
			(function_definition
				declarator: (function_declarator
					declarator: (identifier) @func.name)) @function
			(class_specifier
				name: (type_identifier) @class.name) @class
			(namespace_definition
				name: (identifier) @namespace.name) @namespace
		`,
		"css": `
			(keyframe_block_list
				(keyframe_block 
					(block 
						(declaration 
							(property_name) @property.name)))) @keyframe
			(rule_set
				(selectors
					(class_selector) @class.name)
				(block)) @rule
			(rule_set
				(selectors
					(id_selector) @id.name)
				(block)) @rule
			(media_statement) @media
		`,
		"html": `
			(element
				(start_tag
					(tag_name) @tag.name)) @element
			(script_element) @script
			(style_element) @style
			(element
				(start_tag
					(attribute
						(attribute_name) @attr.name))) @element
		`,
		"lua": `
			(function_declaration
				name: (identifier) @func.name) @function
			(function_definition
				name: (dot_index_expression) @method.name) @method
			(local_function
				name: (identifier) @local_func.name) @local_function
			(table_constructor) @table
		`,
	}

	return queries[fileType]
}

func (a *DefaultAnalyzer) findImports(node *sitter.Node, content []byte, fileType string, lang *sitter.Language) []string {
	var imports []string
	queryStr := a.getImportQueryForLanguage(fileType)
	if queryStr == "" {
		return imports
	}

	query, err := sitter.NewQuery([]byte(queryStr), lang)
	if err != nil {
		return imports
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, node)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		match = qc.FilterPredicates(match, content)
		for _, capture := range match.Captures {
			imports = append(imports, string(capture.Node.Content(content)))
		}
	}

	return imports
}

func (a *DefaultAnalyzer) getImportQueryForLanguage(fileType string) string {
	queries := map[string]string{
		"go": `
			(import_spec 
				path: (interpreted_string_literal) @import)
		`,
		"javascript": `
			(import_statement 
				source: (string) @import)
			(import_from_statement
				source: (string) @import)
		`,
		"python": `
			(import_statement
				name: (dotted_name) @import)
			(import_from_statement
				module_name: (dotted_name) @import)
		`,
		"java": `
			(import_declaration
				name: (identifier) @import)
		`,
		"rust": `
			(use_declaration 
				path: (identifier) @import)
		`,
		"css": `
			(@import_statement
				source: (string_value) @import)
			(@import_statement
				source: (url) @import)
		`,
		"html": `
			(element
				(start_tag
					(attribute
						(attribute_name) @attr
						(quoted_attribute_value) @value)) @tag
				(#eq? @attr "href"))
			(element
				(start_tag
					(attribute
						(attribute_name) @attr
						(quoted_attribute_value) @value)) @tag
				(#eq? @attr "src"))
		`,
		"lua": `
			(function_call
				(identifier) @require
				(arguments
					(string) @import)
				(#eq? @require "require"))
		`,
	}

	return queries[fileType]
}

func (a *DefaultAnalyzer) calculateNodeComplexity(node *sitter.Node) int {
	complexity := 1

	// Count control structures
	cursor := sitter.NewTreeCursor(node)
	defer cursor.Close()

	controlStructures := map[string]bool{
		"if_statement":           true,
		"for_statement":          true,
		"while_statement":        true,
		"switch_statement":       true,
		"catch_clause":           true,
		"conditional_expression": true,
		"binary_expression":      true, // For && and ||
	}

	var traverse func()
	traverse = func() {
		if controlStructures[cursor.CurrentNode().Type()] {
			complexity++
		}

		if cursor.GoToFirstChild() {
			traverse()
			cursor.GoToParent()
		}

		if cursor.GoToNextSibling() {
			traverse()
		}
	}

	traverse()
	return complexity
}

func (a *DefaultAnalyzer) basicAnalyze(path string, content []byte, fileType string) (*FileContext, error) {
	// Provide basic analysis for unsupported file types
	return &FileContext{
		Path:      path,
		Type:      fileType,
		UpdatedAt: time.Now(),
		// Basic complexity based on file size
		Complexity: len(content) / 1000,
	}, nil
}
