package factories

import (
	"context"
	"fmt"
	"time"
	"golang.org/x/time/rate"
	"github.com/jabafett/quill/internal/utils/config"
	"github.com/jabafett/quill/internal/utils/ai"
)

// Provider defines the interface for AI providers
type Provider interface {
	Generate(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error)
}

func (p *rateLimitedProvider) Generate(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error) {
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

// CreateProvider creates a new AI provider instance
func NewProvider(cfg *config.Config) (Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	name := cfg.Core.DefaultProvider

	options, err := config.ConfigToOptions(cfg, name)
	if err != nil {
		return nil, err
	}

	var baseProvider Provider
	switch name {
	case "gemini":
		baseProvider, err = ai.NewGeminiProvider(options)
	case "anthropic":
		baseProvider, err = ai.NewAnthropicProvider(options)
	case "openai":
		baseProvider, err = ai.NewOpenAIProvider(options)
	case "ollama":
		baseProvider, err = ai.NewOllamaProvider(options)
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}

	if err != nil {
		return nil, err
	}

	return &rateLimitedProvider{
		base:          baseProvider,
		enableRetries: options.EnableRetries,
	}, nil
}

const (
	maxRetries = 3
	rateLimit  = time.Second // 1 request per second
)

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

// generateWithRateLimit applies rate limiting to generation requests
func generateWithRateLimit(ctx context.Context, provider Provider, prompt string, opts ai.GenerateOptions) ([]string, error) {
	if err := limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}
	return provider.Generate(ctx, prompt, opts)
}

// rateLimitedProvider wraps a base provider with rate limiting and retry logic
type rateLimitedProvider struct {
	base          Provider
	enableRetries bool
}

// expose rateLimitedProvider for testing
func GetRateLimitedProvider(base Provider, enableRetries bool) *rateLimitedProvider {
	return &rateLimitedProvider{
		base:          base,
		enableRetries: enableRetries,
	}
}