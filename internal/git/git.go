package git

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
)

type Repository struct {
	repo *git.Repository
}

// NewRepository creates a new Repository instance
func NewRepository() (*Repository, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	return &Repository{repo: r}, nil
}

// GetStagedDiff returns the git diff for staged changes
func (r *Repository) GetStagedDiff() (string, error) {
	w, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	var diff strings.Builder
	for path, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified {
			// Get the file's content
			file, err := w.Filesystem.Open(path)
			if err != nil {
				continue // Skip files we can't open
			}
			defer file.Close()

			// Write the diff header
			diff.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", path, path))
			
			// Add status indicators
			switch fileStatus.Staging {
			case git.Added:
				diff.WriteString(fmt.Sprintf("+++ b/%s\n", path))
			case git.Modified:
				diff.WriteString(fmt.Sprintf("--- a/%s\n+++ b/%s\n", path, path))
			case git.Deleted:
				diff.WriteString(fmt.Sprintf("--- a/%s\n", path))
			}
		}
	}

	return diff.String(), nil
}

// IsGitRepo checks if the current directory is a git repository
func IsGitRepo() bool {
	_, err := git.PlainOpen(".")
	return err == nil
}

// HasStagedChanges checks if there are any staged changes
func (r *Repository) HasStagedChanges() (bool, error) {
	w, err := r.repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	for _, s := range status {
		if s.Staging != git.Unmodified {
			return true, nil
		}
	}

	return false, nil
}

// GetChangedFiles returns a list of modified files
func (r *Repository) GetChangedFiles() ([]string, error) {
	w, err := r.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var files []string
	for file, s := range status {
		if s.Staging != git.Unmodified {
			files = append(files, file)
		}
	}

	return files, nil
}

// GetFileType returns the type of changes for a file
func (r *Repository) GetFileType(path string) (string, error) {
	w, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	fileStatus, ok := status[path]
	if !ok {
		return "", fmt.Errorf("file not found in status")
	}

	switch fileStatus.Staging {
	case git.Added:
		return "added", nil
	case git.Modified:
		return "modified", nil
	case git.Deleted:
		return "deleted", nil
	default:
		return "unknown", nil
	}
}

// Commit creates a new git commit with the given message
func (r *Repository) Commit(message string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Create commit
	_, err = w.Commit(message, &git.CommitOptions{})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}