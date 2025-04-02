package context

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jabafett/quill/internal/utils/debug"
)

// RepoSummary contains a simplified summary of the repository
type RepoSummary struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Files       int       `json:"files"`
	Directories int       `json:"directories"`
	Languages   []string  `json:"languages"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SimpleContext provides basic repository context without complex analysis
type SimpleContext struct {
	RepoRoot string
	CachePath string
}

// NewSimpleContext creates a new simple context provider
func NewSimpleContext(repoRoot string) (*SimpleContext, error) {
	if repoRoot == "" {
		return nil, fmt.Errorf("repository root path is required")
	}

	cachePath := filepath.Join(repoRoot, ".git", "quill-summary.json")
	
	return &SimpleContext{
		RepoRoot: repoRoot,
		CachePath: cachePath,
	}, nil
}

// GetDirectoryTree returns a simplified directory tree
func (sc *SimpleContext) GetDirectoryTree(maxDepth int) (string, error) {
	cmd := exec.Command("find", sc.RepoRoot, "-type", "d", "-not", "-path", "*/\\.*", "-maxdepth", fmt.Sprintf("%d", maxDepth))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get directory tree: %w", err)
	}

	// Format the output
	dirs := strings.Split(strings.TrimSpace(string(output)), "\n")
	var tree strings.Builder
	
	for _, dir := range dirs {
		// Skip the root directory
		if dir == sc.RepoRoot {
			continue
		}
		
		// Calculate relative path and indentation
		relPath, err := filepath.Rel(sc.RepoRoot, dir)
		if err != nil {
			continue
		}
		
		depth := len(strings.Split(relPath, string(filepath.Separator))) - 1
		indent := strings.Repeat("  ", depth)
		
		tree.WriteString(fmt.Sprintf("%s- %s\n", indent, filepath.Base(dir)))
	}
	
	return tree.String(), nil
}

// GetFileInfo returns information about files in the repository
func (sc *SimpleContext) GetFileInfo() (map[string]int, error) {
	// Use git ls-files to get all tracked files
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = sc.RepoRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get file list: %w", err)
	}
	
	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	// Count files by extension
	extensions := make(map[string]int)
	for _, file := range files {
		ext := filepath.Ext(file)
		if ext == "" {
			ext = "no_extension"
		}
		extensions[ext]++
	}
	
	return extensions, nil
}

// GetLanguageInfo returns information about programming languages used
func (sc *SimpleContext) GetLanguageInfo() ([]string, error) {
	fileInfo, err := sc.GetFileInfo()
	if err != nil {
		return nil, err
	}
	
	// Map extensions to languages (simplified)
	extToLang := map[string]string{
		".go":   "Go",
		".js":   "JavaScript",
		".ts":   "TypeScript",
		".jsx":  "React",
		".tsx":  "React/TypeScript",
		".py":   "Python",
		".rb":   "Ruby",
		".java": "Java",
		".c":    "C",
		".cpp":  "C++",
		".h":    "C/C++ Header",
		".rs":   "Rust",
		".php":  "PHP",
		".html": "HTML",
		".css":  "CSS",
		".md":   "Markdown",
		".json": "JSON",
		".yml":  "YAML",
		".yaml": "YAML",
		".toml": "TOML",
	}
	
	// Count languages
	langCount := make(map[string]int)
	for ext, count := range fileInfo {
		if lang, ok := extToLang[ext]; ok {
			langCount[lang] += count
		}
	}
	
	// Convert to sorted slice
	var languages []string
	for lang := range langCount {
		languages = append(languages, lang)
	}
	
	return languages, nil
}

// SaveSummary saves the repository summary to the cache file
func (sc *SimpleContext) SaveSummary(summary *RepoSummary) error {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}
	
	err = os.WriteFile(sc.CachePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write summary file: %w", err)
	}
	
	return nil
}

// LoadSummary loads the repository summary from the cache file
func (sc *SimpleContext) LoadSummary() (*RepoSummary, error) {
	data, err := os.ReadFile(sc.CachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read summary file: %w", err)
	}
	
	var summary RepoSummary
	err = json.Unmarshal(data, &summary)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
	}
	
	return &summary, nil
}

// HasSummary checks if a summary file exists
func (sc *SimpleContext) HasSummary() bool {
	_, err := os.Stat(sc.CachePath)
	return err == nil
}

// GetRepoSummary returns a basic summary of the repository
func (sc *SimpleContext) GetRepoSummary() string {
	summary, err := sc.LoadSummary()
	if err != nil {
		debug.Log("Failed to load repository summary: %v", err)
		return ""
	}
	
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Repository: %s\n", summary.Name))
	result.WriteString(fmt.Sprintf("Description: %s\n", summary.Description))
	result.WriteString(fmt.Sprintf("Files: %d\n", summary.Files))
	result.WriteString(fmt.Sprintf("Directories: %d\n", summary.Directories))
	result.WriteString(fmt.Sprintf("Languages: %s\n", strings.Join(summary.Languages, ", ")))
	
	return result.String()
}