package context

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
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
	Analyze(ctx context.Context, path string) (*FileContext, error)
}

// DefaultAnalyzer provides tree-sitter based file analysis
type DefaultAnalyzer struct {
	languages    map[string]*sitter.Language
	queries      map[string]*sitter.Query
	typeDetector *FileTypeDetector
	langLoaders  map[string]func() *sitter.Language
	cursorPool   *sync.Pool
	parserPool   *sync.Pool
	mu           sync.RWMutex
}

// NewDefaultAnalyzer creates a new analyzer with initialized parsers
func NewDefaultAnalyzer() *DefaultAnalyzer {
	a := &DefaultAnalyzer{
		languages:    make(map[string]*sitter.Language),
		queries:      make(map[string]*sitter.Query),
		typeDetector: &FileTypeDetector{},
		langLoaders:  make(map[string]func() *sitter.Language),
		cursorPool: &sync.Pool{
			New: func() interface{} {
				return sitter.NewQueryCursor()
			},
		},
		parserPool: &sync.Pool{
			New: func() interface{} {
				return sitter.NewParser()
			},
		},
	}

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
	// First try read-only access
	a.mu.RLock()
	lang, ok := a.languages[fileType]
	a.mu.RUnlock()

	if ok {
		return lang, nil
	}

	// If not found, acquire write lock
	a.mu.Lock()
	defer a.mu.Unlock()

	// Double-check after acquiring write lock
	if lang, ok := a.languages[fileType]; ok {
		return lang, nil
	}

	loader, ok := a.langLoaders[fileType]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", fileType)
	}

	lang = loader()
	a.languages[fileType] = lang
	return lang, nil
}

func (a *DefaultAnalyzer) Analyze(ctx context.Context, path string) (*FileContext, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	fileType := a.typeDetector.DetectFileType(path, content)

	lang, err := a.getLanguage(fileType)
	if err != nil {
		return a.basicAnalyze(path, content, fileType)
	}

	// Get parser from pool
	parser := a.parserPool.Get().(*sitter.Parser)
	parser.SetLanguage(lang)
	defer a.parserPool.Put(parser)

	tree, err := parser.ParseCtx(ctx, nil, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}
	defer tree.Close()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, fmt.Errorf("failed to get root node")
	}

	symbols, err := a.extractSymbolsFromTree(rootNode, content, fileType, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to extract symbols: %w", err)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	imports, err := a.findImports(rootNode, content, fileType, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to find imports: %w", err)
	}

	complexity := a.calculateNodeComplexity(rootNode)

	return &FileContext{
		Path:       path,
		Type:       fileType,
		Symbols:    symbols,
		Imports:    imports,
		Complexity: complexity,
		UpdatedAt:  time.Now(),
		AST:        rootNode.String(),
	}, nil
}

func (a *DefaultAnalyzer) extractSymbolsFromTree(node *sitter.Node, content []byte, fileType string, lang *sitter.Language) ([]Symbol, error) {
	var symbols []Symbol
	query, err := a.getQuery("symbol", fileType, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to get query: %w", err)
	}

	// Get cursor from pool
	qc := a.cursorPool.Get().(*sitter.QueryCursor)
	defer a.cursorPool.Put(qc)

	qc.Exec(query, node)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			if capture.Node == nil {
				continue
			}

			content := capture.Node.Content(content)
			if len(content) == 0 {
				continue
			}

			symbol := Symbol{
				Name:      string(content),
				Type:      capture.Node.Type(),
				StartLine: int(capture.Node.StartPoint().Row) + 1,
				EndLine:   int(capture.Node.EndPoint().Row) + 1,
			}
			symbol.Complexity = a.calculateNodeComplexity(capture.Node)
			symbols = append(symbols, symbol)
		}
	}

	return symbols, nil
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
            ; Functions
            (function_declaration 
                name: (identifier) @function.name)
            ; Classes
            (class_declaration 
                name: (identifier) @class.name)
            ; Methods
            (method_definition 
                name: (property_identifier) @method.name)
            ; Function expressions in variable declarations
            (variable_declarator
                name: (identifier) @var.name
                value: (function_expression))
            ; Arrow functions in variable declarations
            (variable_declarator
                name: (identifier) @var.name
                value: (arrow_function))
            ; Object methods
            (pair
                key: (property_identifier) @method.name
                value: (function_expression))
            ; Arrow functions in object literals
            (pair
                key: (property_identifier) @method.name
                value: (arrow_function))
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

func (a *DefaultAnalyzer) getQuery(queryType, fileType string, lang *sitter.Language) (*sitter.Query, error) {
	cacheKey := fmt.Sprintf("%s_%s", queryType, fileType)

	a.mu.RLock()
	if query, ok := a.queries[cacheKey]; ok {
		a.mu.RUnlock()
		return query, nil
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	if query, ok := a.queries[cacheKey]; ok {
		return query, nil
	}

	var queryStr string
	switch queryType {
	case "symbol":
		queryStr = a.getSymbolQueryForLanguage(fileType)
	case "import":
		queryStr = a.getImportQueryForLanguage(fileType)
	default:
		return nil, fmt.Errorf("unknown query type: %s", queryType)
	}

	if queryStr == "" {
		return nil, nil
	}

	query, err := sitter.NewQuery([]byte(queryStr), lang)
	if err != nil {
		return nil, fmt.Errorf("failed to create query: %w", err)
	}

	a.queries[cacheKey] = query
	return query, nil
}

func (a *DefaultAnalyzer) findImports(node *sitter.Node, content []byte, fileType string, lang *sitter.Language) ([]string, error) {
	var imports []string
	query, err := a.getQuery("import", fileType, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to get query: %w", err)
	}

	// Get cursor from pool
	qc := a.cursorPool.Get().(*sitter.QueryCursor)
	defer a.cursorPool.Put(qc)

	qc.Exec(query, node)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			if capture.Node == nil {
				continue
			}
			importPath := capture.Node.Content(content)
			if len(importPath) == 0 {
				continue
			}

			// Clean up the import path (remove quotes for Go imports)
			importPath = strings.Trim(string(importPath), `"'`)
			if importPath != "" {
				imports = append(imports, importPath)
			}
		}
	}

	return imports, nil
}

func (a *DefaultAnalyzer) calculateNodeComplexity(node *sitter.Node) int {
	complexity := 1

	// Create a new cursor for each calculation to avoid concurrency issues
	cursor := sitter.NewTreeCursor(node)
	defer cursor.Close()

	controlStructures := map[string]bool{
		"if_statement":           true,
		"for_statement":          true,
		"while_statement":        true,
		"switch_statement":       true,
		"catch_clause":           true,
		"conditional_expression": true,
		"binary_expression":      true,
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

func (a *DefaultAnalyzer) getImportQueryForLanguage(fileType string) string {
	queries := map[string]string{
		"go": `
            ; Standard imports
            (import_declaration 
                (import_spec_list
                    (import_spec
                        path: (interpreted_string_literal) @import.path)))
            ; Single imports
            (import_declaration
                (import_spec
                    path: (interpreted_string_literal) @import.path))
        `,
		"javascript": `
            ; ES6 imports
            (import_statement 
                source: (string) @import.path)
            ; CommonJS require
            (call_expression
                function: (identifier) @require
                arguments: (arguments (string) @import.path)
                (#eq? @require "require"))
        `,
		"python": `
            ; Import statements
            (import_statement 
                name: (dotted_name) @import.path)
            ; From imports
            (import_from_statement 
                module_name: (dotted_name) @import.path)
        `,
		"java": `
            (import_declaration
                name: (identifier) @import.path)
        `,
		"rust": `
            (use_declaration 
                path: (identifier) @import.path)
        `,
		"cpp": `
            (preproc_include
                path: (string_literal) @import.path)
            (preproc_include
                path: (system_lib_string) @import.path)
        `,
	}
	return queries[fileType]
}

// Add cleanup method to properly close resources
func (a *DefaultAnalyzer) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Close all queries
	for _, query := range a.queries {
		if query != nil {
			query.Close()
		}
	}

	// Clear maps
	a.languages = make(map[string]*sitter.Language)
	a.queries = make(map[string]*sitter.Query)

	// Clear pools
	a.parserPool = nil
	a.cursorPool = nil
}
