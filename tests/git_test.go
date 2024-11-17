package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jabafett/quill/internal/git"
)

func TestGitOperations(t *testing.T) {
	// Create temporary git repo
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Initialize git repo
	err = runGitCommand(t, "git", "init")
	if err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Create and add a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = runGitCommand(t, "git", "add", "test.txt")
	if err != nil {
		t.Fatalf("Failed to stage file: %v", err)
	}

	// Test git operations
	repo, err := git.NewRepository()
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Test HasStagedChanges
	hasChanges, err := repo.HasStagedChanges()
	if err != nil {
		t.Fatalf("Failed to check staged changes: %v", err)
	}
	if !hasChanges {
		t.Error("Expected staged changes, got none")
	}

	// Test GetChangedFiles
	files, err := repo.GetChangedFiles()
	if err != nil {
		t.Fatalf("Failed to get changed files: %v", err)
	}
	if len(files) != 1 || files[0] != "test.txt" {
		t.Errorf("Expected ['test.txt'], got %v", files)
	}
}

func runGitCommand(t *testing.T, _ string, _ ...string) error {
	t.Helper()
	// Implementation of git command runner
	return nil // Simplified for example
}
