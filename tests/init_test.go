package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/jabafett/quill/internal/cmd"
)

func TestPromptForProvider(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid gemini selection",
			input:   "1\n",
			want:    "gemini",
			wantErr: false,
		},
		{
			name:    "valid anthropic selection",
			input:   "2\n",
			want:    "anthropic",
			wantErr: false,
		},
		{
			name:    "valid openai selection",
			input:   "3\n",
			want:    "openai",
			wantErr: false,
		},
		{
			name:    "valid ollama selection",
			input:   "4\n",
			want:    "ollama",
			wantErr: false,
		},
		{
			name:        "invalid selection",
			input:       "invalid\n",
			wantErr:     true,
			errContains: "invalid choice",
		},
		{
			name:        "out of range selection",
			input:       "5\n",
			wantErr:     true,
			errContains: "invalid choice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup input
			oldStdin := os.Stdin
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdin = r
			
			// Write test input
			go func() {
				defer w.Close()
				w.Write([]byte(tt.input))
			}()

			// Call function
			got, err := cmd.PromptForProvider()
			
			// Cleanup
			os.Stdin = oldStdin
			r.Close()

			// Check results
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %v", tt.errContains, err)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("PromptForProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     []string
	}{
		{
			name:     "gemini config",
			provider: "gemini",
			want: []string{
				`default_provider = "gemini"`,
				"model = \"gemini-1.5-flash-002\"",
				"max_tokens = 8192",
			},
		},
		{
			name:     "ollama config",
			provider: "ollama",
			want: []string{
				`default_provider = "ollama"`,
				"model = \"qwen2.5-8b-instruct\"",
				"enable_retries = true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cmd.GenerateConfig(tt.provider)
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("GenerateConfig() missing expected content %q", want)
				}
			}
		})
	}
}

func TestGetProviderConfig(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     []string
	}{
		{
			name:     "gemini provider",
			provider: "gemini",
			want: []string{
				"model = \"gemini-1.5-flash-002\"",
				"max_tokens = 8192",
				"temperature = 0.3",
			},
		},
		{
			name:     "invalid provider",
			provider: "invalid",
			want:     []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cmd.GetProviderConfig(tt.provider)
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("GetProviderConfig() missing expected content %q", want)
				}
			}
		})
	}
}
