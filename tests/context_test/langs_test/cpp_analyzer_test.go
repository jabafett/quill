package langs_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jabafett/quill/internal/factories"
	types "github.com/jabafett/quill/internal/utils/context"
	"github.com/jabafett/quill/tests/mocks"
)

func TestCppComplexAnalyzer(t *testing.T) {
	tmpDir := t.TempDir()
	engine, err := factories.NewContextEngine(
		factories.WithMaxConcurrency(1),
		factories.WithCache(true),
		factories.WithCachePath(tmpDir),
	)
	if err != nil {
		t.Fatalf("NewContextEngine() error = %v", err)
	}

	// Create test directory and file
	testDir := t.TempDir()
	cppFile := filepath.Join(testDir, "cpp/advanced.cpp")
	cppContent := `
#include <memory>
#include <map>
#include <string>
#include <optional>

template<typename T>
class Singleton {
protected:
    static std::shared_ptr<T> instance;
    
    Singleton() = default;
    virtual ~Singleton() = default;

public:
    static std::shared_ptr<T> getInstance() {
        if (!instance) {
            instance = std::shared_ptr<T>(new T());
        }
        return instance;
    }
};

template<typename T>
std::shared_ptr<T> Singleton<T>::instance = nullptr;

namespace Services {
    template<typename T>
    class IService {
    public:
        virtual ~IService() = default;
        virtual std::optional<T> process(const T& data) = 0;
    };

    template<typename T>
    class ComplexService : public IService<T>, public Singleton<ComplexService<T>> {
    private:
        std::map<std::string, T> cache;
        
    public:
        std::optional<T> process(const T& data) override {
            // Implementation
            return std::nullopt;
        }
        
        std::optional<T> cachedCompute(const std::string& key) {
            auto it = cache.find(key);
            if (it != cache.end()) {
                return it->second;
            }
            return std::nullopt;
        }
    };

    enum class LogLevel {
        DEBUG,
        INFO,
        WARNING,
        ERROR
    };

    struct Configuration {
        LogLevel level;
        std::string path;
    };
}`

	// Write test file
	if err := mocks.EnsureParentDir(cppFile); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}
	if err := os.WriteFile(cppFile, []byte(cppContent), 0644); err != nil {
		t.Fatalf("Failed to write advanced.cpp: %v", err)
	}

	// Extract context
	ctx, err := engine.ExtractContext([]string{cppFile}, false)
	if err != nil {
		t.Fatalf("ExtractContext() error = %v", err)
	}

	// Get the file context
	fileCtx, ok := ctx.Files[cppFile]
	require.True(t, ok, "File context not found for %s", cppFile)

	// Define expected symbols
	symbols := []struct {
		name string
		typ  string
	}{
		{"Singleton", string(types.Class)},
		{"IService", string(types.Class)},
		{"ComplexService", string(types.Class)},
		{"LogLevel", string(types.Enum)},
		{"Configuration", string(types.Class)},
		{"process", string(types.Function)},
		{"cachedCompute", string(types.Function)},
		{"Services", string(types.Class)},
		{"DEBUG", string(types.Constant)},
		{"INFO", string(types.Constant)},
		{"WARNING", string(types.Constant)},
		{"ERROR", string(types.Constant)},
	}

	// Verify all expected symbols are present
	for _, expected := range symbols {
		found := false
		for _, symbol := range fileCtx.Symbols {
			if symbol.Name == expected.name && string(symbol.Type) == expected.typ {
				found = true
				break
			}
		}
		assert.True(t, found, "Symbol %s of type %s not found", expected.name, expected.typ)
	}

	if t.Failed() {
		jsonCtx, err := json.MarshalIndent(ctx, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal context: %v", err)
		}
		t.Logf("Context:\n%s", string(jsonCtx))
	}

}
