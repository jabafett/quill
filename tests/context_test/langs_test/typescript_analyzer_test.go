package langs_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jabafett/quill/internal/factories"
	types "github.com/jabafett/quill/internal/utils/context"
	"github.com/jabafett/quill/tests/mocks"
)

func TestTypeScriptComplexAnalyzer(t *testing.T) {
	tmpDir := t.TempDir()
	engine, err := factories.NewContextEngine(
		factories.WithMaxConcurrency(1),
		factories.WithCache(true),
		factories.WithCachePath(tmpDir),
	)
	require.NoError(t, err)

	// Create test directory and file
	testDir := t.TempDir()
	tsFile := filepath.Join(testDir, "src/service.ts")
	tsContent := `
import { Observable, BehaviorSubject } from 'rxjs';
import { map, filter, catchError } from 'rxjs/operators';

interface DataHandler<T> {
    process(data: T): Promise<T>;
    validate(data: T): boolean;
}

type ProcessorConfig = {
    retries: number;
    timeout: number;
    validateBeforeProcess: boolean;
};

abstract class BaseProcessor<T> implements DataHandler<T> {
    protected config: ProcessorConfig;
    private statusSubject: BehaviorSubject<string>;

    constructor(config: ProcessorConfig) {
        this.config = config;
        this.statusSubject = new BehaviorSubject<string>('idle');
    }

    public get status$(): Observable<string> {
        return this.statusSubject.asObservable();
    }

    abstract process(data: T): Promise<T>;

    public validate(data: T): boolean {
        return true;
    }

    protected async withRetry<R>(
        operation: () => Promise<R>
    ): Promise<R> {
        let lastError: Error | null = null;
        
        for (let attempt = 1; attempt <= this.config.retries; attempt++) {
            try {
                return await operation();
            } catch (error) {
                lastError = error as Error;
                this.statusSubject.next('Retry failed');
            }
        }
        
        throw lastError;
    }
}`

	// Write test file
	if err := mocks.EnsureParentDir(tsFile); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}
	if err := os.WriteFile(tsFile, []byte(tsContent), 0644); err != nil {
		t.Fatalf("Failed to write service.ts: %v", err)
	}

	// Extract context
	ctx, err := engine.ExtractContext([]string{tsFile}, false)
	if err != nil {
		t.Fatalf("ExtractContext() error = %v", err)
	}

	// Get the file context
	fileCtx, ok := ctx.Files[tsFile]
	require.True(t, ok, "File context not found for %s", tsFile)

	// Define expected symbols
	symbols := []struct {
		name string
		typ  string
	}{
		{"DataHandler", string(types.Interface)},
		{"ProcessorConfig", string(types.Type)},
		{"BaseProcessor", string(types.Class)},
		{"process", string(types.Function)},
		{"validate", string(types.Function)},
		{"withRetry", string(types.Function)},
		{"status$", string(types.Function)},
		{"statusSubject", string(types.Field)},
	}

	// Test all symbols in a loop
	for _, sym := range symbols {
		mocks.AssertSymbol(t, fileCtx.Symbols, sym.name, sym.typ)
	}

	// Verify imports are correctly detected
	assert.Contains(t, fileCtx.Imports, "rxjs")
	assert.Contains(t, fileCtx.Imports, "rxjs/operators")

	if t.Failed() {
		jsonCtx, err := json.MarshalIndent(ctx, "", "  ")
		if err != nil {
			t.Errorf("Failed to marshal JSON: %v", err)
		}
		fmt.Println(string(jsonCtx))
	}
}
