package context

import "time"

// FIRST IMPLEMENTATION

// Context represents the full codebase context
type Context struct {
	Files      map[string]*FileContext
	UpdatedAt  time.Time
	Complexity int
	References map[string][]string
	Errors     []Error // Track analysis errors
}

// FileContext represents context for a single file
type FileContext struct {
	Path       string
	Type       string
	Symbols    []Symbol
	Imports    []string
	Complexity int
	UpdatedAt  time.Time
	AST        string  // Store AST representation
	Errors     []Error // Store parsing errors
}

// Symbol represents a code symbol (function, class, etc.)
type Symbol struct {
	Name       string
	Type       string
	StartLine  int
	EndLine    int
	Complexity int
	Children   []Symbol               // Support nested symbols
	Metadata   map[string]interface{} // Store additional symbol metadata
}

// SECOND IMPLEMENTATION

// Repository represents the main structure of the repository context schema.
type Repository struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	RepositoryURL *string `json:"repositoryUrl,omitempty"`
	Visibility    string  `json:"visibility"` // "public" | "private"

	VersionControl struct {
		Branch           string `json:"branch"`
		LastCommitHeader string `json:"lastCommitHeader"`
		LastTagHeader    string `json:"lastTagHeader"`
	} `json:"versionControl"`

	Metrics struct {
		TotalFiles   int      `json:"totalFiles"`
		TotalLines   int      `json:"totalLines"`
		TestCoverage *float64 `json:"testCoverage,omitempty"`
	} `json:"metrics"`

	Languages struct {
		Primary string   `json:"primary"`
		Others  []string `json:"others"`
	} `json:"languages"`

	Frameworks struct {
		Frontend []string `json:"frontend,omitempty"`
		Backend  []string `json:"backend,omitempty"`
		Testing  []string `json:"testing,omitempty"`
		Build    []string `json:"build,omitempty"`
	} `json:"frameworks"`

	DependenciesSummary struct {
		Total    int          `json:"total"`
		Critical []Dependency `json:"critical"`
	} `json:"dependenciesSummary"`

	Architecture struct {
		Pattern    string      `json:"pattern"`
		Components []Component `json:"components"`
		DataFlow   []string    `json:"dataFlow"`
	} `json:"architecture"`

	Files          []FileInfo `json:"files"`
	IsListComplete bool       `json:"isListComplete"`
	Errors         []Error    `json:"errors,omitempty"`
}

// Dependency represents a single dependency in the repository.
type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// File represents a file in the repository with associated metadata
type File struct {
	FileInfo        FileInfo  `json:"fileInfo"`
	IsComplete      bool      `json:"isComplete"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Errors          []Error   `json:"errors,omitempty"`
	IsTest          bool      `json:"isTest"`
	IsConfig        bool      `json:"isConfig"`
	IsDocumentation bool      `json:"isDocumentation"`
	IsFileComplete  bool      `json:"isFileComplete"`
}

// FileInfo represents detailed information about a file
type FileInfo struct {
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	Extension string   `json:"extension"`
	MimeType  string   `json:"mimeType"`
	Imports   []string `json:"imports"`
	Symbols   []Symbol `json:"symbols"`
}

// Symbol represents a symbol found in a file with associated metadata
//type Symbol struct {
//	SymbolInfo SymbolInfo `json:"symbolInfo"`
//	Errors     []Error    `json:"errors,omitempty"`
//}

// SymbolInfo represents information about a symbol
type SymbolInfo struct {
	Type          string   `json:"type"`
	Name          string   `json:"name"`
	Implements    []string `json:"implements,omitempty"`
	Documentation *string  `json:"documentation,omitempty"` // cap at 100
	AST           *string  `json:"ast,omitempty"`
	Complexity    *int     `json:"complexity,omitempty"`
}

// Component represents a component in the repository architecture.
type Component struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Responsibility string   `json:"responsibility"`
	Dependencies   []string `json:"dependencies"`
	APIs           []API    `json:"apis,omitempty"`
}

// API represents an API endpoint exposed by a component.
type API struct {
	Path     string      `json:"path"`
	Method   string      `json:"method"`
	Auth     bool        `json:"auth"`
	Params   []Parameter `json:"params,omitempty"`
	Response string      `json:"response"`
}

// Parameter represents a parameter in an API endpoint.
type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// Error represents a parsing or analysis error
type Error struct {
	Path    string
	Line    int
	Column  int
	Message string
}
