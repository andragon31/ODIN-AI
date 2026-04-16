package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// OpenRouterProvider is a provider for OpenRouter models
type OpenRouterProvider struct {
	config OpenRouterConfig
	client *http.Client
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider(config OpenRouterConfig) *OpenRouterProvider {
	if config.Endpoint == "" {
		config.Endpoint = DefaultOpenRouterEndpoint
	}

	return &OpenRouterProvider{
		config: config,
		client: &http.Client{
			Timeout: Timeout,
		},
	}
}

// Name returns the provider name
func (p *OpenRouterProvider) Name() string {
	return "openrouter"
}

// Supports checks if the provider supports a given model
func (p *OpenRouterProvider) Supports(model string) bool {
	// OpenRouter supports many models - we assume all standard models are supported
	// In production, you would check against a list of available models
	supportedModels := map[string]bool{
		"anthropic/claude-3.5-sonnet": true,
		"anthropic/claude-3-opus":     true,
		"openai/gpt-4-turbo":          true,
		"openai/gpt-4":                true,
		"google/gemini-pro-1.5":       true,
		"meta-llama/llama-3-70b":      true,
		"mistralai/mistral-7b":        true,
	}

	// If model starts with any known prefix, it's supported
	for known := range supportedModels {
		if len(model) >= len(known) && model[:len(known)] == known {
			return true
		}
	}

	// Allow any model if API key is set (OpenRouter will error if invalid)
	return p.config.APIKey != ""
}

// Generate generates text from the model
func (p *OpenRouterProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	if req.MaxTokens == 0 {
		req.MaxTokens = 2048
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	// Convert messages to OpenRouter format
	openRouterReq := map[string]interface{}{
		"model": req.Model,
		"messages": func() []map[string]string {
			msgs := make([]map[string]string, len(req.Messages))
			for i, msg := range req.Messages {
				msgs[i] = map[string]string{
					"role":    msg.Role,
					"content": msg.Content,
				}
			}
			return msgs
		}(),
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	}

	jsonData, err := json.Marshal(openRouterReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		p.config.Endpoint+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	httpReq.Header.Set("HTTP-Referer", "https://odin-ai.dev")
	httpReq.Header.Set("X-Title", "ODIN AI")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("openrouter error: %s (code %d)", errResp.Error.Message, errResp.Error.Code)
	}

	var openRouterResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openRouterResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openrouter")
	}

	// Calculate cost based on model
	cost := p.CostPerToken(req.Model) * float64(openRouterResp.Usage.TotalTokens)

	return &GenerateResponse{
		Content: openRouterResp.Choices[0].Message.Content,
		Usage: Usage{
			InputTokens:  openRouterResp.Usage.PromptTokens,
			OutputTokens: openRouterResp.Usage.CompletionTokens,
			TotalTokens:  openRouterResp.Usage.TotalTokens,
			Cost:         cost,
		},
		Model: req.Model,
	}, nil
}

// Embed generates embeddings (not supported by OpenRouter directly)
func (p *OpenRouterProvider) Embed(ctx context.Context, texts []string) ([]Embedding, error) {
	// OpenRouter doesn't support embeddings directly
	// Fall back to a simple hash-based embedding for compatibility
	embeddings := make([]Embedding, len(texts))
	for i, text := range texts {
		// Simple hash-based pseudo-embedding
		vec := make([]float64, 384)
		h := hashStrings(text)
		for j := range vec {
			vec[j] = float64((h>>uint(j))&0xFF) / 255.0
		}
		embeddings[i] = Embedding{
			Vector:  vec,
			Content: text,
		}
	}
	return embeddings, nil
}

// hashStrings creates a simple hash of a string
func hashStrings(s string) uint64 {
	var h uint64
	for i, c := range s {
		h ^= uint64(c) + 0x9e3779b97f4a7c15 + (h << 6) + (h >> 2)
		h *= uint64(i + 1)
	}
	return h
}

// CostPerToken returns the cost per token for OpenRouter models
// Based on https://openrouter.ai/models pricing (approximate)
func (p *OpenRouterProvider) CostPerToken(model string) float64 {
	costs := map[string]float64{
		"anthropic/claude-3.5-sonnet": 0.000003,   // $3/1M input, $15/1M output
		"anthropic/claude-3-opus":     0.000015,   // $15/1M input, $75/1M output
		"openai/gpt-4-turbo":          0.00001,    // $10/1M input, $30/1M output
		"openai/gpt-4":                0.00003,    // $30/1M input, $60/1M output
		"google/gemini-pro-1.5":       0.00000125, // $1.25/1M input, $5/1M output
		"meta-llama/llama-3-70b":      0.0000008,  // $0.80/1M
		"mistralai/mistral-7b":        0.00000024, // $0.24/1M
	}

	if cost, ok := costs[model]; ok {
		return cost
	}

	// Default fallback cost
	return 0.000001
}

// IsAvailable checks if OpenRouter API is available
func (p *OpenRouterProvider) IsAvailable(ctx context.Context) bool {
	if p.config.APIKey == "" {
		return false
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		p.config.Endpoint+"/models", nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
