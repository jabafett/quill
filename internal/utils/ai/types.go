package ai

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