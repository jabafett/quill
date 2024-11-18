package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jabafett/quill/internal/utils/git"
)

func TestGitOperations(t *testing.T) {
	// Create temporary git repo
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Set git config
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Create and add a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
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
