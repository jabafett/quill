package factories

import (
	"context"
	"fmt"
	"time"

	"github.com/jabafett/quill/internal/utils/ai"
	"github.com/jabafett/quill/internal/utils/config"
	"golang.org/x/time/rate"
)

// Provider defines the interface for AI providers
type Provider interface {
	Generate(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error)
}

// rateLimitedProvider wraps a base provider with rate limiting and retry logic
type rateLimitedProvider struct {
	base          Provider
	enableRetries bool
}

// GenerateOptions contains options for the generate factory
type ProviderOptions struct {
	Provider    string
	Candidates  int
	Temperature float32
}

func (p *rateLimitedProvider) Generate(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error) {
	if p.enableRetries {
		var result []string
		err := retryWithBackoff(ctx, func() error {
			var genErr error
			result, genErr = generateSingleInstance(ctx, p.base, prompt, opts)
			return genErr
		})
		return result, err
	}

	return generateSingleInstance(ctx, p.base, prompt, opts)
}

// CreateProvider creates a new AI provider instance
func NewProvider(cfg *config.Config, opts ProviderOptions) (Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	name := cfg.Core.DefaultProvider

	options, err := config.ConfigToOptions(cfg, name)
	if err != nil {
		return nil, err
	}

	if opts.Candidates > 0 {
		options.CandidateCount = opts.Candidates
	}
	if opts.Temperature > 0 {
		options.Temperature = opts.Temperature
	}
	if opts.Provider != "" {
		name = opts.Provider
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

// generateSingleInstance ensures only one request is made at a time
func generateSingleInstance(ctx context.Context, provider Provider, prompt string, opts ai.GenerateOptions) ([]string, error) {
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

// ProviderOptions contains options for provider factories
type ProviderOptions struct {
        Provider    string
        Candidates  int
        Temperature float32
        StagedOnly  bool // Only consider staged changes (for suggest command)
        UnstagedOnly bool // Only consider unstaged changes (for suggest command)
}

func (p *rateLimitedProvider) Generate(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error) {
        if p.enableRetries {
                var result []string
                err := retryWithBackoff(ctx, func() error {
                        var genErr error
                        result, genErr = generateSingleInstance(ctx, p.base, prompt, opts)
                        return genErr
                })
                return result, err
        }

        return generateSingleInstance(ctx, p.base, prompt, opts)
}

// CreateProvider creates a new AI provider instance
func NewProvider(cfg *config.Config, opts ProviderOptions) (Provider, error) {
        if cfg == nil {
                return nil, fmt.Errorf("config cannot be nil")
        }

        name := cfg.Core.DefaultProvider

        options, err := config.ConfigToOptions(cfg, name)
        if err != nil {
                return nil, err
        }

        if opts.Candidates > 0 {
                options.CandidateCount = opts.Candidates
        }
        if opts.Temperature > 0 {
                options.Temperature = opts.Temperature
        }
        if opts.Provider != "" {
                name = opts.Provider
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

// generateSingleInstance ensures only one request is made at a time
func generateSingleInstance(ctx context.Context, provider Provider, prompt string, opts ai.GenerateOptions) ([]string, error) {
        if err := limiter.Wait(ctx); err != nil {
                return nil, fmt.Errorf("rate limit wait failed: %w", err)
        }
        return provider.Generate(ctx, prompt, opts)
}


// expose rateLimitedProvider for testing
func GetRateLimitedProvider(base Provider, enableRetries bool) *rateLimitedProvider {
        return &rateLimitedProvider{
                base:          base,
                enableRetries: enableRetries,
        }
}
