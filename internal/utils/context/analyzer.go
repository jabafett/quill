package context

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jabafett/quill/internal/utils/debug"
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
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
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

	// Track language statistics
	languageStats map[string]int
	totalFiles    int
	totalLines    int
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
		languageStats: make(map[string]int),
	}
	a.registerLanguages()
	return a
}

// registerLanguages registers supported languages
func (a *DefaultAnalyzer) registerLanguages() {
	a.langLoaders = map[string]func() *sitter.Language{
		"go":         golang.GetLanguage,
		"javascript": javascript.GetLanguage,
		"typescript": typescript.GetLanguage,
		"tsx":        tsx.GetLanguage,
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

// getLanguage returns the language for the given file type
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

// Analyzes a single file
func (a *DefaultAnalyzer) Analyze(ctx context.Context, path string) (*FileContext, error) {
	// Get file info first to capture mod time
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	modTime := fileInfo.ModTime()

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	fileType := a.typeDetector.DetectFileType(path, content)

	// Update language statistics with thread safety
	a.mu.Lock()
	if fileType != "text/plain" && fileType != "" {
		a.languageStats[fileType]++
	}
	a.totalFiles++
	a.totalLines += len(strings.Split(string(content), "\n"))
	a.mu.Unlock()

	lang, err := a.getLanguage(fileType)
	if err != nil {
		// Pass modTime to basicAnalyze
		return a.basicAnalyze(path, fileType, modTime)
	}

	// Get parser from pool and ensure it's properly reset
	parser := a.parserPool.Get().(*sitter.Parser)
	parser.SetLanguage(lang)

	// Create a copy of the content for this parse operation
	contentCopy := make([]byte, len(content))
	copy(contentCopy, content)

	// Parse with the copied content
	tree, err := parser.ParseCtx(ctx, nil, contentCopy)
	if err != nil {
		a.parserPool.Put(parser)
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Get root node before putting parser back
	rootNode := tree.RootNode()
	if rootNode == nil {
		tree.Close()
		a.parserPool.Put(parser)
		return nil, fmt.Errorf("failed to get root node")
	}

	// Create a new cursor for this analysis
	cursor := sitter.NewTreeCursor(rootNode)
	defer cursor.Close()

	// Extract all necessary information while we still have the tree
	var errors []Error
	if rootNode.HasError() {
		errors = append(errors, Error{
			Message: fmt.Sprintf("Syntax error in file %s", path),
		})
	}

	symbols, err := a.extractSymbolsFromTree(rootNode, contentCopy, fileType, lang)
	if err != nil {
		errors = append(errors, Error{
			Message: fmt.Sprintf("Failed to extract symbols: %v", err),
		})
	}

	imports, err := a.findImports(rootNode, contentCopy, fileType, lang)
	if err != nil {
		errors = append(errors, Error{
			Message: fmt.Sprintf("Failed to find imports: %v", err),
		})
	}

	// Clean up tree-sitter resources
	tree.Close()
	a.parserPool.Put(parser)

	// UpdatedAt should reflect the time of this analysis run.
	updatedAt := time.Now()
	// Keep the actual file mod time separate in ModTime field.

	fileCtx := &FileContext{
		Path:      path,
		Type:      fileType,
		Symbols:   symbols,
		Imports:   imports,
		UpdatedAt: updatedAt, // Use the analysis time
		ModTime:   modTime,   // Keep the actual file mod time separate
		Errors:    errors,
	}

	debug.Log("Analyzer.Analyze: Created FileContext for %s with UpdatedAt: %s, ModTime: %s", path, fileCtx.UpdatedAt.Format(time.RFC3339Nano), fileCtx.ModTime.Format(time.RFC3339Nano)) // Log both

	return fileCtx, nil
}

// extractSymbolsFromTree extracts symbols using tree-sitter queries based on capture names.
func (a *DefaultAnalyzer) extractSymbolsFromTree(node *sitter.Node, content []byte, fileType string, lang *sitter.Language) ([]SymbolContext, error) {
	var symbols []SymbolContext
	if node == nil || len(content) == 0 || lang == nil {
		return symbols, nil
	}

	query, err := a.getQuery("symbol", fileType, lang)
	if err != nil {
		// If no symbol query exists for the language, return empty symbols gracefully
		if strings.Contains(err.Error(), "no query available") {
			return symbols, nil
		}
		return nil, fmt.Errorf("failed to get symbol query: %w", err)
	}

	// Get cursor from pool
	qc := a.cursorPool.Get().(*sitter.QueryCursor)
	defer a.cursorPool.Put(qc) // Return cursor to pool when done

	qc.Exec(query, node)

	// Map capture names to SymbolType
	captureToSymbolType := map[string]SymbolType{
		"func.name":        Function,
		"method.name":      Function,
		"getter.name":      Function,
		"setter.name":      Function,
		"constructor.name": Function,
		"class.name":       Class,
		"struct.name":      Class, // Treat structs like classes
		"interface.name":   Interface,
		"trait.name":       Interface, // Treat traits like interfaces
		"enum.name":        Enum,
		"type.name":        Type,
		"var.name":         Variable,
		"const.name":       Constant,
		"field.name":       Field,
		"property.name":    Field, // Treat properties like fields
		"ivar.name":        Field,
		"cvar.name":        Field,
		"module.name":      Module,
		"annotation.name":  Modifier,
		// TS/TSX specific captures
		"component.name":      Constant,
		"react_component":     Constant,
		"function.name":       Function,
		"method_signature":    Function,
		"call_signature":      Function,
		"arrow_function":      Function,
		"type_alias":          Type,
		"public_field":        Field,
		"property_identifier": Field,
	}

	// Process matches
	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		var symbolName string
		var symbolType SymbolType
		var symbolNode *sitter.Node

		// Find the primary capture for name and determine type
		for _, capture := range match.Captures {
			captureName := query.CaptureNameForId(capture.Index)
			if st, ok := captureToSymbolType[captureName]; ok {
				symbolType = st
				symbolNode = capture.Node // Node containing the symbol's name/identifier

				// Extract the name content safely
				if symbolNode != nil {
					startByte := symbolNode.StartByte()
					endByte := symbolNode.EndByte()
					if startByte < uint32(len(content)) && endByte <= uint32(len(content)) && startByte < endByte {
						symbolName = string(content[startByte:endByte])
					}
				}
				break // Found the primary name capture for this match
			}
		}

		// If we found a valid symbol name and type, create the context
		if symbolName != "" && symbolType != "" && symbolNode != nil {
			// Use the node of the *entire match* for line numbers, not just the name node
			matchNode := match.Captures[0].Node // Assuming first capture covers the whole pattern
			if len(match.Captures) > 0 && match.Captures[0].Node != nil {
				matchNode = match.Captures[0].Node
			}

			symbolCtx := SymbolContext{
				Name:      symbolName,
				Type:      string(symbolType),
				StartLine: int(matchNode.StartPoint().Row) + 1,
				EndLine:   int(matchNode.EndPoint().Row) + 1,
			}
			symbols = append(symbols, symbolCtx)
		}
	}

	return symbols, nil
}

// findImports returns a list of import paths from the given node
func (a *DefaultAnalyzer) findImports(node *sitter.Node, content []byte, fileType string, lang *sitter.Language) ([]string, error) {
	var imports []string
	if node == nil || len(content) == 0 || lang == nil {
		return imports, nil
	}

	query, err := a.getQuery("import", fileType, lang)
	if err != nil {
		// Gracefully handle missing import queries
		if strings.Contains(err.Error(), "no query available") {
			return imports, nil
		}
		return nil, fmt.Errorf("failed to get import query: %w", err)
	}

	// Get cursor from pool
	qc := a.cursorPool.Get().(*sitter.QueryCursor)
	defer a.cursorPool.Put(qc) // Return cursor to pool

	qc.Exec(query, node)

	// Process matches
	processedPaths := make(map[string]bool) // Track unique import paths per file
	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		// Find the capture named "@import.path"
		for _, capture := range match.Captures {
			captureName := query.CaptureNameForId(capture.Index)
			if captureName == "import.path" && capture.Node != nil {
				// Extract the path content safely
				node := capture.Node
				startByte := node.StartByte()
				endByte := node.EndByte()
				if startByte < uint32(len(content)) && endByte <= uint32(len(content)) && startByte < endByte {
					importPath := string(content[startByte:endByte])
					// Basic cleaning (remove quotes) - normalization happens later
					importPath = strings.Trim(importPath, `"'`)
					// Skip non-package imports like "Data service initialized"
					if importPath != "" && !processedPaths[importPath] && !strings.Contains(importPath, " ") {
						imports = append(imports, importPath)
						processedPaths[importPath] = true
					}
				}
				break // Found the path for this match
			}
		}
	}
	return imports, nil
}

// basicAnalyze provides basic analysis for unsupported file types
// Accepts modTime from the caller to ensure consistency
func (a *DefaultAnalyzer) basicAnalyze(path string, fileType string, modTime time.Time) (*FileContext, error) {
	// Provide basic analysis for unsupported file types
	return &FileContext{
		Path:      path,
		Type:      fileType,
		UpdatedAt: time.Now(), // Analysis time
		ModTime:   modTime,    // File system mod time
	}, nil
}

// Clear pools
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

	// Clear pools - we need to create new empty pools rather than setting to nil
	a.parserPool = &sync.Pool{
		New: func() interface{} {
			return sitter.NewParser()
		},
	}
	a.cursorPool = &sync.Pool{
		New: func() interface{} {
			return sitter.NewQueryCursor()
		},
	}
}

// GetLanguages returns the primary (most frequently used) language and secondary languages
func (a *DefaultAnalyzer) GetLanguages() (string, []string) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var maxCount int
	var primary string
	var others []string

	for lang, count := range a.languageStats {
		if count > maxCount {
			maxCount = count
			primary = lang
		}
		if lang != "text/plain" && lang != "" {
			others = append(others, lang)
		}
	}

	// Remove primary from others if it exists
	for i, lang := range others {
		if lang == primary {
			others = append(others[:i], others[i+1:]...)
			break
		}
	}

	return primary, others
}

// GetTotalFiles returns the total number of files analyzed
func (a *DefaultAnalyzer) GetTotalFiles() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.totalFiles
}

// GetTotalLines returns the total number of lines analyzed
func (a *DefaultAnalyzer) GetTotalLines() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.totalLines
}

// GetLanguageStats returns a copy of the language statistics map
func (a *DefaultAnalyzer) GetLanguageStats() map[string]int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats := make(map[string]int, len(a.languageStats))
	for k, v := range a.languageStats {
		stats[k] = v
	}
	return stats
}

// AddLanguageStats adds to the language statistics count
func (a *DefaultAnalyzer) AddLanguageStats(lang string, count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.languageStats[lang] += count
}

// AddTotalFiles adds to the total files count
func (a *DefaultAnalyzer) AddTotalFiles(count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.totalFiles += count
}

// AddTotalLines adds to the total lines count
func (a *DefaultAnalyzer) AddTotalLines(count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.totalLines += count
}
