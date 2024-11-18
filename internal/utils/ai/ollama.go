package ai

import (
    "context"
    "fmt"
	"bytes"
	"encoding/json"
	"net/http"
)

type OllamaProvider struct {
	options Options
	baseURL string
}

type ollamaRequest struct {
    Model       string  `json:"model"`
    Prompt      string  `json:"prompt"`
    Temperature float32 `json:"temperature"`
}

type ollamaResponse struct {
    Response string `json:"response"`
}

func NewOllamaProvider(options Options) (*OllamaProvider, error) {
	return &OllamaProvider{
		options: options,
        baseURL: "http://localhost:11434/api/generate",
    }, nil
}

func (p *OllamaProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) ([]string, error) {
	maxCandidates := opts.MaxCandidates
	if maxCandidates <= 0 {
        maxCandidates = p.options.CandidateCount
    }
    if maxCandidates > 3 {
        maxCandidates = 3
    }

    temperature := p.options.Temperature
    if opts.Temperature != nil {
        temperature = *opts.Temperature
    }

    reqBody := ollamaRequest{
        Model:       p.options.Model,
        Prompt:      prompt,
        Temperature: temperature,
    }

    var responses []string
    for i := 0; i < maxCandidates; i++ {
        resp, err := p.makeRequest(ctx, reqBody)
        if err != nil {
            return responses, fmt.Errorf("failed to generate candidate %d: %w", i+1, err)
        }
        responses = append(responses, resp)
    }

    return responses, nil
}

func (p *OllamaProvider) makeRequest(ctx context.Context, reqBody ollamaRequest) (string, error) {
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return "", fmt.Errorf("failed to marshal request: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    var result ollamaResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("failed to decode response: %w", err)
    }

    return result.Response, nil
} 