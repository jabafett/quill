package context

import "time"

// Context represents the full codebase context
type Context struct {
	Files       map[string]*FileContext
	UpdatedAt   time.Time
	Complexity  int
	References  map[string][]string
	Errors      []Error // Track analysis errors
}

// FileContext represents context for a single file
type FileContext struct {
	Path       string
	Type       string
	Symbols    []Symbol
	Imports    []string
	Complexity int
	UpdatedAt  time.Time
	AST        string    // Store AST representation
	Errors     []Error   // Store parsing errors
}

// Symbol represents a code symbol (function, class, etc.)
type Symbol struct {
	Name       string
	Type       string
	StartLine  int
	EndLine    int
	Complexity int
	Children   []Symbol // Support nested symbols
	Metadata   map[string]interface{} // Store additional symbol metadata
}

// Error represents a parsing or analysis error
type Error struct {
	Path    string
	Line    int
	Column  int
	Message string
}
  