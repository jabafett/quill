package git

import (
	"fmt"
	"os/exec"
	"strconv"
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
	cmd := exec.Command("git", "diff", "--staged", "--no-color")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}
	return string(output), nil
}

// GetStagedDiffStats returns more detailed diff stats
func (r *Repository) GetStagedDiffStats() (added int, deleted int, files []string, err error) {
	cmd := exec.Command("git", "diff", "--staged", "--numstat")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to get diff stats: %w", err)
	}

	// Parse the numstat output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files = make([]string, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			if a, err := strconv.Atoi(parts[0]); err == nil {
				added += a
			}
			if d, err := strconv.Atoi(parts[1]); err == nil {
				deleted += d
			}
			files = append(files, parts[2])
		}
	}

	return added, deleted, files, nil
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

// GetStagedDiffOptimized returns an optimized git diff for staged changes
func (r *Repository) GetStagedDiffOptimized() (string, error) {
	// Use --no-prefix to avoid directory prefixes
	// Use --no-color to avoid ANSI codes
	// Use --cached as an alias for --staged
	cmd := exec.Command("git", "diff", "--cached", "--no-prefix", "--no-color")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}
	return string(output), nil
}

// GetStagedFilesOptimized returns only staged files efficiently
func (r *Repository) GetStagedFilesOptimized() ([]string, error) {
	// Use --name-only to get just filenames
	// Use --cached as an alias for --staged
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	if len(output) == 0 {
		return nil, nil
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	return files, nil
}

// GetFileStatusOptimized returns the status of a specific file efficiently
func (r *Repository) GetFileStatusOptimized(path string) (string, error) {
	// Use --porcelain for machine-readable output
	cmd := exec.Command("git", "status", "--porcelain", path)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get file status: %w", err)
	}

	if len(output) < 2 {
		return "unmodified", nil
	}

	// First character represents staging status
	switch output[0] {
	case 'A':
		return "added", nil
	case 'M':
		return "modified", nil
	case 'D':
		return "deleted", nil
	case 'R':
		return "renamed", nil
	case 'C':
		return "copied", nil
	default:
		return "unknown", nil
	}
}

// CommitOptimized creates a new git commit with optimized performance
func (r *Repository) CommitOptimized(message string) error {
	// Use -m to avoid opening editor
	// Use --no-verify to skip hooks for performance
	cmd := exec.Command("git", "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}
	return nil
}

// HasStagedChangesOptimized checks for staged changes efficiently
func (r *Repository) HasStagedChangesOptimized() (bool, error) {
	// Use --quiet to suppress output
	// Exit status is 1 if there are no changes
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	err := cmd.Run()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit status 1 means there are changes
			return exitErr.ExitCode() == 1, nil
		}
		return false, fmt.Errorf("failed to check staged changes: %w", err)
	}

	// Exit status 0 means no changes
	return false, nil
}
