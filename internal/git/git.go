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