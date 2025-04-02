package context

import "time"

// Context represents the main structure of the repository context schema.
type RepositoryContext struct {
	Name          string                  `json:"name,omitempty"`
	Description   string                  `json:"description,omitempty"` // ONLY ai filled field, created last with the available context
	Files         map[string]*FileContext `json:"files"`
	RepositoryURL string                  `json:"repositoryUrl,omitempty"`
	Visibility    string                  `json:"visibility,omitempty"`   // "public" | "private"
	Dependencies  []Dependency            `json:"dependencies,omitempty"` // any dependency imported into a file

	VersionControl VersionControl `json:"versionControl,omitempty"`
	Metrics        Metrics        `json:"metrics,omitempty"`
	Languages      Languages      `json:"languages,omitempty"`

	// dev fields
	IsListComplete bool    `json:"isListComplete"`
	Errors         []Error `json:"errors,omitempty"`
}

// FileContext represents context for a single file
type FileContext struct {
	Path      string          `json:"path"`
	Type      string          `json:"type"`
	Symbols   []SymbolContext `json:"symbols,omitempty"`
	Imports   []string        `json:"imports,omitempty"`
	UpdatedAt time.Time       `json:"updatedAt"` // Time the context was last updated/analyzed
	ModTime   time.Time       `json:"modTime"`   // File system modification time when analyzed
	Errors    []Error         `json:"errors,omitempty"`
}

// Symbol represents a code symbol (function, class, etc.)
type SymbolContext struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	FilePath  string `json:"filePath,omitempty"`  // File containing the symbol
	Error     Error  `json:"error,omitempty"`
	// Track relationships
	CalledBy  []int64 `json:"calledBy,omitempty"`  // SymbolIDs that call this symbol
	Variables []int64 `json:"variables,omitempty"` // SymbolIDs that are used by this symbol
}

// Dependency represents a single dependency in the repository.
type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"-"`
}

// Error represents a parsing or analysis error
type Error struct {
	Message string `json:"message,omitempty"`
}

type SymbolType string

const (
	Function  SymbolType = "function"
	Class     SymbolType = "class"
	Interface SymbolType = "interface"
	Type      SymbolType = "type"
	Field     SymbolType = "field"
	Enum      SymbolType = "enum"
	Module    SymbolType = "module"
	Constant  SymbolType = "constant"
	Variable  SymbolType = "variable" // Added Variable type
	Modifier  SymbolType = "modifier"
)

type VersionControl struct {
	Branch           string `json:"branch,omitempty"`
	LastCommitHeader string `json:"lastCommitHeader,omitempty"`
	LastCommitDate   string `json:"lastCommitDate,omitempty"`
	LastTagHeader    string `json:"lastTagHeader,omitempty"`
	LastTagDate      string `json:"lastTagDate,omitempty"`
}

type Metrics struct {
	TotalFiles   int      `json:"totalFiles"`
	TotalLines   int      `json:"totalLines"`
	TestCoverage *float64 `json:"testCoverage,omitempty"`
}

type Languages struct {
	Primary string   `json:"primary"`
	Others  []string `json:"others"`
}
