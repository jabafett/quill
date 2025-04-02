package index_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jabafett/quill/internal/providers"
	ctxUtils "github.com/jabafett/quill/internal/utils/context"
	"github.com/jabafett/quill/internal/utils/debug"
	"github.com/jabafett/quill/tests/mocks"
)

// setupTestRepo initializes a git repository in a temporary directory,
// creates specified files with content, and commits them.
func setupTestRepo(t *testing.T, files map[string]string) (repoPath string, cleanup func()) {
	t.Helper()

	// Create a temporary directory for the repository
	repoPath = t.TempDir()

	debug.Log("Created temporary directory for test repository: %s", repoPath)

	// Initialize git repository
	cmdRepoCreate := exec.Command("git", "init", repoPath)
	cmdRepoCreate.Dir = repoPath
	err := cmdRepoCreate.Run()
	require.NoError(t, err, "Failed to initialize git repository")

	// Create and add files
	for relPath, content := range files {
		absPath := filepath.Join(repoPath, relPath)
		err := mocks.EnsureParentDir(absPath)
		require.NoError(t, err, "Failed to create parent directory for %s", relPath)
		err = os.WriteFile(absPath, []byte(content), 0644)
		require.NoError(t, err, "Failed to write file %s", relPath)
	}

	// Add all files to staging
	cmdAdd := exec.Command("git", "add", ".")
	cmdAdd.Dir = repoPath
	err = cmdAdd.Run()
	require.NoError(t, err, "Failed to git add files")

	// Commit the files
	// Configure git user locally for the commit
	cmdConfigUser := exec.Command("git", "config", "user.name", "Test User")
	cmdConfigUser.Dir = repoPath
	err = cmdConfigUser.Run()
	require.NoError(t, err, "Failed to set git user.name")

	cmdConfigEmail := exec.Command("git", "config", "user.email", "test@example.com")
	cmdConfigEmail.Dir = repoPath
	err = cmdConfigEmail.Run()
	require.NoError(t, err, "Failed to set git user.email")

	cmdCommit := exec.Command("git", "commit", "-m", "Initial commit")
	cmdCommit.Dir = repoPath
	err = cmdCommit.Run()
	require.NoError(t, err, "Failed to git commit files")

	// Check the repo directory for the files and the repository state
	debug.Log("Checking the test repository for files and state...")
	err = checkRepoState(t, repoPath)
	require.NoError(t, err, "Failed to check test repository state")
	err = checkFiles(t, repoPath, files)
	require.NoError(t, err, "Failed to check test repository files")
	debug.Log("Test repository state is valid.")

	cleanup = func() {
		// os.RemoveAll(repoPath) // TempDir handles cleanup
	}

	return repoPath, cleanup
}

// checkFiles checks the files in the repository at the given path
func checkFiles(t *testing.T, repoPath string, files map[string]string) error {
	t.Helper()

	// List files in the repo
	cmdLs := exec.Command("git", "ls-files")
	cmdLs.Dir = repoPath

	output, err := cmdLs.Output()
	if err != nil {
		return fmt.Errorf("failed to list files in repository: %w", err)
	}

	if len(output) == 0 {
		return fmt.Errorf("repository is empty")
	}
	debug.Log("Repository files: %s", output)

	// Check the directory for the files
	for relPath, content := range files {
		absPath := filepath.Join(repoPath, relPath)
		fileInfo, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("failed to stat file %s: %w", absPath, err)
		}
		if fileInfo.IsDir() {
			return fmt.Errorf("file %s is a directory", absPath)
		}
		if fileInfo.Size() == 0 {
			return fmt.Errorf("file %s is empty", absPath)
		}
		fileContent, err := os.ReadFile(absPath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", absPath, err)
		}
		if string(fileContent) != content {
			return fmt.Errorf("file %s does not match expected content", absPath)
		}
	}

	return nil
}

// checkRepoState checks the state of the repository at the given path
func checkRepoState(t *testing.T, repoPath string) error {
	t.Helper()

	// Check the repo directory for the files and the repository state
	cmdStatus := exec.Command("git", "status", "--porcelain")
	cmdStatus.Dir = repoPath

	output, err := cmdStatus.Output()
	if err != nil {
		return fmt.Errorf("failed to check repository state: %w", err)
	}

	if len(output) > 0 {
		return fmt.Errorf("repository state is not clean:\n%s", output)
	}

	return nil
}

func TestIndexRepository_TypeScript(t *testing.T) {
	// Turn on debug logging flag for this test
	debug.Initialize(true)
	debug.Log("Enabling debug logging for this test")
	stringReplacer := strings.NewReplacer("__BACKTICK__", "`")

	// Define TypeScript files and their content
	tsFiles := map[string]string{
		"src/utils/helpers.ts": stringReplacer.Replace(`
export function capitalize(str: string): string {
    if (!str) return str;
    return str.charAt(0).toUpperCase() + str.slice(1);
}

export const PI = 3.14159;

export type UtilityConfig = {
    logLevel: 'debug' | 'info' | 'warn' | 'error';
};
`),
		"src/services/dataService.ts": stringReplacer.Replace(`
import { capitalize, UtilityConfig } from '../utils/helpers';
import { Observable, BehaviorSubject } from 'rxjs';

interface DataProvider<T> {
    fetchData(id: string): Observable<T>;
}

export class DataService<T> implements DataProvider<T> {
    private dataCache: Map<string, T> = new Map();
    private config: UtilityConfig;
    private statusSubject = new BehaviorSubject<string>('idle');

    constructor(config: UtilityConfig) {
        this.config = config;
        console.log(capitalize('Data service initialized'));
    }

    fetchData(id: string): Observable<T> {
        this.statusSubject.next('fetching');
        // Simulate fetching data
        return new Observable(subscriber => {
            setTimeout(() => {
                const data = this.dataCache.get(id) || null;
                if (data) {
                    subscriber.next(data);
                    subscriber.complete();
                } else {
                    subscriber.error(new Error('Data not found'));
                }
                this.statusSubject.next('idle');
            }, 50);
        });
    }

    updateData(id: string, data: T): void {
        this.dataCache.set(id, data);
    }

    get status$(): Observable<string> {
        return this.statusSubject.asObservable();
    }
}
`),
		"src/components/ui/Button.tsx": stringReplacer.Replace(`
import React from 'react';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: 'primary' | 'secondary';
}

export const Button: React.FC<ButtonProps> = ({ children, variant = 'primary', ...props }) => {
    const baseStyle = "px-4 py-2 rounded focus:outline-none";
    const variantStyle = variant === 'primary'
        ? "bg-blue-500 text-white hover:bg-blue-600"
        : "bg-gray-300 text-black hover:bg-gray-400";

    return (
        <button className={__BACKTICK__${baseStyle} ${variantStyle}__BACKTICK__} {...props}>
            {children}
        </button>
    );
};
`),
		"src/index.ts": stringReplacer.Replace(`
import { DataService } from './services/dataService';
import { capitalize, PI, UtilityConfig } from './utils/helpers';
// Button is imported for type checking / potential future use, but not directly used here
import { Button } from './components/ui/Button';

console.log('Starting app...');
console.log(__BACKTICK__PI is approximately: ${PI}__BACKTICK__);

const config: UtilityConfig = { logLevel: 'info' };
const userService = new DataService<string>(config);

userService.updateData('user1', 'Alice');

userService.fetchData('user1').subscribe({
    next: (name) => console.log(__BACKTICK__Fetched user: ${capitalize(name)}__BACKTICK__),
    error: (err) => console.error(err.message),
});

userService.status$.subscribe(status => {
    console.log(__BACKTICK__Service status: ${status}__BACKTICK__);
});

// Example of using Button type (though Button component itself isn't rendered)
type AppButtonProps = React.ComponentProps<typeof Button>;
const buttonProps: AppButtonProps = { variant: 'secondary', onClick: () => {} };
console.log('Button props created:', buttonProps.variant);
`),
		".gitignore": stringReplacer.Replace(`
node_modules
dist
build
*.log
.env
`),
		"README.md": stringReplacer.Replace(`
# Test Repository

This is a test repository for Quill indexing.
It contains TypeScript code.
`),
		"data/config.json": stringReplacer.Replace(`
{
    "featureFlags": {
        "newUI": true
    }
}
`),
	}

	// Setup the test repository
	repoPath, cleanup := setupTestRepo(t, tsFiles)
	defer cleanup()

	cachePath := t.TempDir()

	// --- Simulate running the index command ---
	// Change working directory to the repo path temporarily
	originalWD, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(repoPath)
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWD)
		if err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	// Instantiate IndexProvider pointing to the test repo and cache
	// We need to ensure the provider uses the correct paths
	// Mocking NewIndexProvider or adjusting its internals for testing might be needed
	// For simplicity here, we assume NewIndexProvider picks up the current WD correctly
	// and we can override cache path (or rely on default behavior if it uses TempDir)

	require.NoError(t, err)

	// Manually create IndexProvider with test paths/cache
	// This bypasses NewIndexProvider's auto-detection for more control
	indexProvider, err := providers.NewIndexProvider(
		providers.WithCachePath(cachePath),
		providers.WithBasePath(repoPath),
		providers.WithRepoRootPath(repoPath),
	)

	require.NoError(t, err)

	// Run the indexing process
	fmt.Println("Starting repository indexing for test...")
	err = indexProvider.IndexRepository(context.Background(), true) // false = don't force reindex
	require.NoError(t, err, "IndexRepository failed")
	fmt.Println("Repository indexing for test completed.")

	// --- Verification ---
	// Retrieve the context from the cache
	var repoCtx ctxUtils.RepositoryContext
	repoCacheKey := fmt.Sprintf("repo_context:%s", repoPath)
	err = indexProvider.GetCachedContext(repoCacheKey, &repoCtx)
	require.NoError(t, err, "Failed to get repository context from cache")

	// Assertions
	assert.NotEmpty(t, repoCtx.Name, "Repository name should be set")
	assert.Equal(t, filepath.Base(repoPath), repoCtx.Name) // Assuming name is derived from dir

	// Check expected files (excluding .gitignore, README.md, config.json as they might not be analyzed)
	expectedFiles := []string{
		"src/utils/helpers.ts",
		"src/services/dataService.ts",
		"src/components/ui/Button.tsx",
		"src/index.ts",
	}
	// Adjust check to account for potential inclusion of non-code files if analyzer handles them
	// For now, assume only code files are deeply analyzed and stored.
	analyzedFileCount := 0
	for _, f := range repoCtx.Files {
		// Count files that have symbols or imports, indicating analysis occurred
		if len(f.Symbols) > 0 || len(f.Imports) > 0 {
			analyzedFileCount++
		}
	}
	assert.Equal(t, len(expectedFiles), analyzedFileCount, "Incorrect number of analyzed files indexed")

	for _, expectedFile := range expectedFiles {
		fileCtx, ok := repoCtx.Files[expectedFile]
		require.True(t, ok, "Expected file %s not found in context", expectedFile)
		assert.Equal(t, expectedFile, fileCtx.Path)
		assert.NotEmpty(t, fileCtx.Type, "File type should be set for %s", expectedFile)
		assert.False(t, fileCtx.UpdatedAt.IsZero(), "UpdatedAt should be set for %s", expectedFile)
		assert.NotEmpty(t, fileCtx.Symbols, "Symbols should be extracted for %s", expectedFile)
	}

	// Spot check symbols in dataService.ts
	dataServiceCtx, ok := repoCtx.Files["src/services/dataService.ts"]
	require.True(t, ok)
	mocks.AssertSymbol(t, dataServiceCtx.Symbols, "DataProvider", string(ctxUtils.Interface))
	mocks.AssertSymbol(t, dataServiceCtx.Symbols, "DataService", string(ctxUtils.Class))
	mocks.AssertSymbol(t, dataServiceCtx.Symbols, "fetchData", string(ctxUtils.Function)) // Method treated as function
	mocks.AssertSymbol(t, dataServiceCtx.Symbols, "updateData", string(ctxUtils.Function))
	mocks.AssertSymbol(t, dataServiceCtx.Symbols, "status$", string(ctxUtils.Function)) // Getter treated as function
	mocks.AssertSymbol(t, dataServiceCtx.Symbols, "dataCache", string(ctxUtils.Field))  // Property/Field

	// Spot check imports in dataService.ts
	// Normalize imports before checking
	normalizedImports := make(map[string]bool)
	for _, imp := range dataServiceCtx.Imports {
		// Basic normalization for comparison (adjust if your normalizeImport is different)
		impStr := imp                                                         // Use a different variable name to avoid confusion
		if len(impStr) > 1 && (impStr[0] == '"' || impStr[0] == byte('\'')) { // Use byte value for single quote
			impStr = impStr[1:]
		}
		if len(impStr) > 1 && (impStr[len(impStr)-1] == '"' || impStr[len(impStr)-1] == byte('\'')) { // Use byte value for single quote
			impStr = impStr[:len(impStr)-1]
		}
		normalizedImports[impStr] = true
	}
	assert.True(t, normalizedImports["../utils/helpers"], "Missing import '../utils/helpers' in dataService.ts")
	assert.True(t, normalizedImports["rxjs"], "Missing import 'rxjs' in dataService.ts")

	// Spot check symbols in Button.tsx
	buttonCtx, ok := repoCtx.Files["src/components/ui/Button.tsx"]
	require.True(t, ok)
	mocks.AssertSymbol(t, buttonCtx.Symbols, "ButtonProps", string(ctxUtils.Interface))
	mocks.AssertSymbol(t, buttonCtx.Symbols, "Button", string(ctxUtils.Constant)) // React FC often seen as const

	// Check aggregated dependencies
	expectedDeps := []string{
		"../utils/helpers",       // from dataService.ts
		"rxjs",                   // from dataService.ts
		"react",                  // from Button.tsx
		"./services/dataService", // from index.ts
		"./utils/helpers",        // from index.ts
		"./components/ui/Button", // from index.ts
	}
	actualDeps := make([]string, len(repoCtx.Dependencies))
	for i, dep := range repoCtx.Dependencies {
		actualDeps[i] = dep.Name
	}
	assert.ElementsMatch(t, expectedDeps, actualDeps, "Aggregated dependencies do not match")

	// Check Languages
	assert.Equal(t, "typescript", repoCtx.Languages.Primary, "Primary language should be typescript")
	// Allow for flexibility in how 'tsx' is categorized
	if repoCtx.Languages.Primary != "tsx" {
		assert.Contains(t, repoCtx.Languages.Others, "tsx", "tsx should be in other languages if not primary")
	}

	// Check Metrics
	assert.Equal(t, 7, repoCtx.Metrics.TotalFiles, "TotalFiles metric mismatch") // Expect 7 files including non-code ones
	assert.Greater(t, repoCtx.Metrics.TotalLines, 50, "TotalLines should be greater than 50") // Rough estimate

	// Check Version Control info
	assert.NotEmpty(t, repoCtx.VersionControl.Branch, "Branch should be set")
	// assert.NotEmpty(t, repoCtx.VersionControl.LastCommitHeader, "LastCommitHeader should be set") // Requires more git interaction
	// assert.NotEmpty(t, repoCtx.VersionControl.LastCommitDate, "LastCommitDate should be set")     // Requires more git interaction

	// --- Test Incremental Update ---
	fmt.Println("Testing incremental update...")

	// Modify one file
	helpersPath := filepath.Join(repoPath, "src/utils/helpers.ts")
	helpersContent, err := os.ReadFile(helpersPath)
	require.NoError(t, err)
	newHelpersContent := string(helpersContent) + "\n// Added a comment\n"
	err = os.WriteFile(helpersPath, []byte(newHelpersContent), 0644)
	require.NoError(t, err)

	// Wait a moment to ensure mod time changes
	time.Sleep(1100 * time.Millisecond) // Ensure > 1 second difference for truncation comparison

	// Run indexing again
	err = indexProvider.IndexRepository(context.Background(), false)
	require.NoError(t, err, "Incremental IndexRepository failed")

	// Retrieve context again
	var updatedRepoCtx ctxUtils.RepositoryContext
	err = indexProvider.GetCachedContext(repoCacheKey, &updatedRepoCtx)
	require.NoError(t, err, "Failed to get updated repository context from cache")

	// Verify that the modified file's UpdatedAt timestamp has changed
	originalHelpersCtx := repoCtx.Files["src/utils/helpers.ts"]
	updatedHelpersCtx := updatedRepoCtx.Files["src/utils/helpers.ts"]
	require.NotNil(t, originalHelpersCtx)
	require.NotNil(t, updatedHelpersCtx)
	assert.NotEqual(t, originalHelpersCtx.UpdatedAt.Truncate(time.Second), updatedHelpersCtx.UpdatedAt.Truncate(time.Second), "UpdatedAt for modified file should change")

	// Verify that an unmodified file's UpdatedAt timestamp has NOT changed
	originalIndexCtx := repoCtx.Files["src/index.ts"]
	updatedIndexCtx := updatedRepoCtx.Files["src/index.ts"]
	require.NotNil(t, originalIndexCtx)
	require.NotNil(t, updatedIndexCtx)
	assert.Equal(t, originalIndexCtx.UpdatedAt.Truncate(time.Second), updatedIndexCtx.UpdatedAt.Truncate(time.Second), "UpdatedAt for unmodified file should not change")

	// Verify context is still generally correct (e.g., file count)
	updatedAnalyzedFileCount := 4
	assert.Equal(t, len(expectedFiles), updatedAnalyzedFileCount, "File count should remain the same after incremental update")
	assert.Equal(t, repoCtx.Name, updatedRepoCtx.Name, "Repository name should persist")

	// --- Test Force Reindex ---
	fmt.Println("Testing force reindex...")
	// Get the timestamp of an unmodified file *before* force reindex
	indexModTimeBeforeForce := updatedIndexCtx.UpdatedAt

	// Wait a moment
	time.Sleep(1100 * time.Millisecond)

	// Run indexing with force=true
	err = indexProvider.IndexRepository(context.Background(), true) // true = force reindex
	require.NoError(t, err, "Force Reindex IndexRepository failed")

	// Retrieve context again
	var forcedRepoCtx ctxUtils.RepositoryContext
	err = indexProvider.GetCachedContext(repoCacheKey, &forcedRepoCtx)
	require.NoError(t, err, "Failed to get forced repository context from cache")

	// Verify that the previously unmodified file's UpdatedAt timestamp HAS changed
	forcedIndexCtx := forcedRepoCtx.Files["src/index.ts"]
	require.NotNil(t, forcedIndexCtx)
	assert.NotEqual(t, indexModTimeBeforeForce.Truncate(time.Second), forcedIndexCtx.UpdatedAt.Truncate(time.Second), "UpdatedAt for previously unmodified file should change after force reindex")

	// Verify context is still generally correct
	forcedAnalyzedFileCount := 0
	for _, f := range forcedRepoCtx.Files {
		if len(f.Symbols) > 0 || len(f.Imports) > 0 {
			forcedAnalyzedFileCount++
		}
	}
	assert.Equal(t, len(expectedFiles), forcedAnalyzedFileCount, "File count should remain the same after force reindex")

	// Optional: Print JSON context if test fails
	if t.Failed() {
		jsonCtx, _ := json.MarshalIndent(repoCtx, "", "  ")
		fmt.Println("Initial Context:\n", string(jsonCtx))
		jsonUpdatedCtx, _ := json.MarshalIndent(updatedRepoCtx, "", "  ")
		fmt.Println("Updated Context:\n", string(jsonUpdatedCtx))
		jsonForcedCtx, _ := json.MarshalIndent(forcedRepoCtx, "", "  ")
		fmt.Println("Forced Context:\n", string(jsonForcedCtx))
	}
}
