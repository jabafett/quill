package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/jabafett/quill/internal/config"
	"github.com/jabafett/quill/internal/keyring"
	"golang.org/x/time/rate"

)

const (
	maxRetries = 3
	rateLimit  = time.Second // 1 request per second
)

// Provider defines the interface for AI providers
type Provider interface {
	Generate(ctx context.Context, prompt string, opts GenerateOptions) ([]string, error)
}

// Options contains common configuration options for AI providers
type Options struct {
	Model          string
	MaxTokens      int
	Temperature    float32
	APIKey         string
	EnableRetries  bool
	CandidateCount int
}

// GenerateOptions contains options for a single generation request
type GenerateOptions struct {
	MaxCandidates int      // Number of variations to generate (max 3)
	MaxTokens     int      // Override default max tokens if needed
	Temperature   *float32 // Override default temperature if needed
}

var (
	// Global rate limiter shared across all providers
	limiter = rate.NewLimiter(rate.Every(rateLimit), 1)
)

// retryWithBackoff implements exponential backoff retry logic
func retryWithBackoff(ctx context.Context, fn func() error) error {
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if err = fn(); err == nil {
			return nil
		}

		// Calculate backoff duration: 2^attempt * base duration
		backoff := time.Duration(1<<uint(attempt)) * time.Second

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			continue
		}
	}
	return fmt.Errorf("max retries exceeded: %w", err)
}

// ConfigToOptions converts a provider config to Options
func ConfigToOptions(cfg *config.Config, providerName string) (Options, error) {
	provider, exists := cfg.Providers[providerName]
	if !exists {
		return Options{}, fmt.Errorf("provider '%s' not found in configuration", providerName)
	}

	// Get API key from keyring
	var kp keyring.Provider
	switch providerName {
	case "gemini":
		kp = keyring.Gemini
	case "anthropic":
		kp = keyring.Anthropic
	case "openai":
		kp = keyring.OpenAI
	default:
		return Options{}, fmt.Errorf("unknown provider: %s", providerName)
	}

	apiKey, err := keyring.GetAPIKey(kp)
	if err != nil {
		return Options{}, fmt.Errorf("failed to get API key: %w", err)
	}

	return Options{
		Model:          provider.Model,
		MaxTokens:      provider.MaxTokens,
		Temperature:    provider.Temperature,
		APIKey:         apiKey,
		EnableRetries:  provider.EnableRetries,
		CandidateCount: provider.CandidateCount,
	}, nil
}

// generateWithRateLimit applies rate limiting to generation requests
func generateWithRateLimit(ctx context.Context, provider Provider, prompt string, opts GenerateOptions) ([]string, error) {
	if err := limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}
	return provider.Generate(ctx, prompt, opts)
}

func NewProvider(name string, cfg *config.Config) (Provider, error) {
	options, err := ConfigToOptions(cfg, name)
	if err != nil {
		return nil, err
	}

	var baseProvider Provider
	// Create provider instance
	switch name {
	case "gemini":
		baseProvider, err = NewGeminiProvider(options)
	case "anthropic":
		baseProvider, err = NewAnthropicProvider(options)
	case "openai":
		baseProvider, err = NewOpenAIProvider(options)
	case "ollama":
		baseProvider, err = NewOllamaProvider(options)
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}

	if err != nil {
		return nil, err
	}

	// Wrap the provider with rate limiting and optional retry logic
	return &rateLimitedProvider{
		base:          baseProvider,
		enableRetries: options.EnableRetries,
	}, nil
}

// rateLimitedProvider wraps a base provider with rate limiting and retry logic
type rateLimitedProvider struct {
	base          Provider
	enableRetries bool
}

func (p *rateLimitedProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) ([]string, error) {
	if p.enableRetries {
		var result []string
		err := retryWithBackoff(ctx, func() error {
			var genErr error
			result, genErr = generateWithRateLimit(ctx, p.base, prompt, opts)
			return genErr
		})
		return result, err
	}

	return generateWithRateLimit(ctx, p.base, prompt, opts)
}

// expose rateLimitedProvider for testing
func GetRateLimitedProvider(base Provider, enableRetries bool) *rateLimitedProvider {
	return &rateLimitedProvider{
		base:          base,
		enableRetries: enableRetries,
	}
}
