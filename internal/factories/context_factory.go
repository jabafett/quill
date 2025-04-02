// Simplified context factory for repository information
package factories

import (
        "fmt"
        "os/exec"
        "path/filepath"
        "strings"

        "github.com/jabafett/quill/internal/utils/context"
        "github.com/jabafett/quill/internal/utils/debug"
)

// ContextOptions contains configuration for the context provider
type ContextOptions struct {
        RepoRootPath string // Absolute path to the repository root
}

// ContextProvider manages basic repository context
type ContextProvider struct {
        simpleContext *context.SimpleContext
        options       ContextOptions
}

// NewContextProvider creates a new context provider instance
func NewContextProvider(opts ...func(*ContextOptions)) (*ContextProvider, error) {
        // Default options
        options := ContextOptions{
                RepoRootPath: "",
        }

        // Apply option overrides
        for _, opt := range opts {
                opt(&options)
        }

        if options.RepoRootPath == "" {
                return nil, fmt.Errorf("repository root path is required")
        }

        simpleContext, err := context.NewSimpleContext(options.RepoRootPath)
        if err != nil {
                return nil, fmt.Errorf("failed to create simple context: %w", err)
        }

        return &ContextProvider{
                simpleContext: simpleContext,
                options:       options,
        }, nil
}

// GetDirectoryTree returns a simplified directory tree
func (p *ContextProvider) GetDirectoryTree(maxDepth int) (string, error) {
        return p.simpleContext.GetDirectoryTree(maxDepth)
}

// GetFileInfo returns information about files in the repository
func (p *ContextProvider) GetFileInfo() (map[string]int, error) {
        return p.simpleContext.GetFileInfo()
}

// GetLanguageInfo returns information about programming languages used
func (p *ContextProvider) GetLanguageInfo() ([]string, error) {
        return p.simpleContext.GetLanguageInfo()
}

// HasSummary checks if a repository summary exists
func (p *ContextProvider) HasSummary() bool {
        return p.simpleContext.HasSummary()
}

// GetRepoSummary returns the repository summary
func (p *ContextProvider) GetRepoSummary() string {
        return p.simpleContext.GetRepoSummary()
}

// SaveSummary saves a repository summary
func (p *ContextProvider) SaveSummary(summary *context.RepoSummary) error {
        return p.simpleContext.SaveSummary(summary)
}

// LoadSummary loads the repository summary
func (p *ContextProvider) LoadSummary() (*context.RepoSummary, error) {
        return p.simpleContext.LoadSummary()
}

// GetRepoDescription generates a description of the repository
func (p *ContextProvider) GetRepoDescription() (string, error) {
        // Get repository name
        cmd := exec.Command("basename", p.options.RepoRootPath)
        output, err := cmd.Output()
        if err != nil {
                return "", fmt.Errorf("failed to get repository name: %w", err)
        }
        repoName := strings.TrimSpace(string(output))

        // Get README content if available
        var readmeContent string
        readmePaths := []string{
                filepath.Join(p.options.RepoRootPath, "README.md"),
                filepath.Join(p.options.RepoRootPath, "README"),
                filepath.Join(p.options.RepoRootPath, "readme.md"),
        }

        for _, path := range readmePaths {
                cmd = exec.Command("cat", path)
                output, err = cmd.Output()
                if err == nil {
                        readmeContent = string(output)
                        break
                }
        }

        // Get file count
        fileInfo, err := p.GetFileInfo()
        if err != nil {
                return "", fmt.Errorf("failed to get file info: %w", err)
        }

        fileCount := 0
        for _, count := range fileInfo {
                fileCount += count
        }

        // Build description
        var description strings.Builder
        description.WriteString(fmt.Sprintf("Repository: %s\n", repoName))
        description.WriteString(fmt.Sprintf("Files: %d\n", fileCount))

        // Add languages if available
        languages, err := p.GetLanguageInfo()
        if err == nil && len(languages) > 0 {
                description.WriteString(fmt.Sprintf("Languages: %s\n", strings.Join(languages, ", ")))
        }

        // Add README excerpt if available
        if readmeContent != "" {
                // Limit to first few lines
                lines := strings.Split(readmeContent, "\n")
                var excerpt []string
                for i, line := range lines {
                        if i >= 5 || len(excerpt) >= 3 { // Take up to 5 lines, but only 3 non-empty ones
                                break
                        }
                        if strings.TrimSpace(line) != "" {
                                excerpt = append(excerpt, strings.TrimSpace(line))
                        }
                }
                if len(excerpt) > 0 {
                        description.WriteString("\nDescription from README:\n")
                        for _, line := range excerpt {
                                description.WriteString(line + "\n")
                        }
                }
        }

        return description.String(), nil
}

// WithRepoRootPath sets the repository root path
func WithRepoRootPath(path string) func(*ContextOptions) {
        return func(opts *ContextOptions) {
                opts.RepoRootPath = path
        }
}
