package tests

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jabafett/quill/internal/ai"
	"github.com/jabafett/quill/tests/mocks"
)

const (
	testRateLimitDuration = 2 * time.Second
	testRetryAttempts    = 3
	testPrompt           = "test prompt"
)

var errTemporary = errors.New("temporary error")

func TestRateLimiting(t *testing.T) {
	mock := &mocks.MockGeminiProvider{
		GenerateFunc: func(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error) {
			if prompt != testPrompt {
				t.Errorf("Expected prompt %q, got %q", testPrompt, prompt)
			}
			return []string{"test"}, nil
		},
	}

	provider := ai.GetRateLimitedProvider(mock, false)

	ctx := context.Background()
	start := time.Now()

	for i := 0; i < testRetryAttempts; i++ {
		_, err := provider.Generate(ctx, testPrompt, ai.GenerateOptions{})
		if err != nil {
			t.Fatalf("Generate failed on attempt %d: %v", i+1, err)
		}
	}

	duration := time.Since(start)
	if duration < testRateLimitDuration {
		t.Errorf("Rate limiting not enforced: took %v, expected at least %v", 
			duration, testRateLimitDuration)
	}
}

func TestRetryBehavior(t *testing.T) {
	var retryCount int32
	mock := &mocks.MockGeminiProvider{
		GenerateFunc: func(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error) {
			currentCount := atomic.AddInt32(&retryCount, 1)
			if currentCount < 3 {
				return nil, errTemporary
			}
			return []string{"success"}, nil
		},
	}

	provider := ai.GetRateLimitedProvider(mock, true)

	ctx := context.Background()
	result, err := provider.Generate(ctx, testPrompt, ai.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if retryCount != 3 { // Initial attempt + 2 retries
		t.Errorf("Expected 3 attempts, got %d", retryCount)
	}

	if len(result) != 1 || result[0] != "success" {
		t.Errorf("Expected [success], got %v", result)
	}
}

func TestRetryDisabled(t *testing.T) {
	var retryCount int32
	mock := &mocks.MockGeminiProvider{
		GenerateFunc: func(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error) {
			atomic.AddInt32(&retryCount, 1)
			return nil, errors.New("error")
		},
	}

	provider := ai.GetRateLimitedProvider(mock, false)

	ctx := context.Background()
	_, err := provider.Generate(ctx, testPrompt, ai.GenerateOptions{})
	if err == nil {
		t.Error("Expected error when retries disabled")
	}

	if retryCount != 1 {
		t.Errorf("Expected single attempt, got %d", retryCount)
	}
}
