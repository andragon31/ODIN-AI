package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AnthropicProvider is a provider for Anthropic models
type AnthropicProvider struct {
	config AnthropicConfig
	client *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(config AnthropicConfig) *AnthropicProvider {
	if config.Endpoint == "" {
		config.Endpoint = DefaultAnthropicEndpoint
	}

	return &AnthropicProvider{
		config: config,
		client: &http.Client{
			Timeout: Timeout,
		},
	}
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// Supports checks if the provider supports a given model
func (p *AnthropicProvider) Supports(model string) bool {
	supportedModels := map[string]bool{
		"claude-3-5-sonnet-20241022": true,
		"claude-3-5-sonnet":          true,
		"claude-3-opus-20240229":     true,
		"claude-3-opus":              true,
		"claude-3-sonnet-20240229":   true,
		"claude-3-sonnet":            true,
		"claude-3-haiku-20240307":    true,
		"claude-3-haiku":             true,
	}
	return supportedModels[model]
}

// Generate generates text from the model
func (p *AnthropicProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	if req.MaxTokens == 0 {
		req.MaxTokens = 2048
	}
	if req.Temperature == 0 {
		req.Temperature = 1.0
	}

	// Convert messages to Anthropic format
	anthropicReq := map[string]interface{}{
		"model":       req.Model,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	}

	// Convert messages - Anthropic uses a different format
	var anthropicMessages []map[string]string
	for _, msg := range req.Messages {
		anthropicMessages = append(anthropicMessages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	anthropicReq["messages"] = anthropicMessages

	jsonData, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		p.config.Endpoint+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("anthropic-dangerous-direct-browser-access", "true")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("anthropic error: %s", errResp.Error.Message)
	}

	var anthropicResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("no response from anthropic")
	}

	// Calculate cost
	cost := p.CostPerToken(req.Model) * float64(anthropicResp.Usage.OutputTokens)

	return &GenerateResponse{
		Content: anthropicResp.Content[0].Text,
		Usage: Usage{
			InputTokens:  anthropicResp.Usage.InputTokens,
			OutputTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:  anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
			Cost:         cost,
		},
		Model: req.Model,
	}, nil
}

// Embed generates embeddings (not supported by Anthropic)
func (p *AnthropicProvider) Embed(ctx context.Context, texts []string) ([]Embedding, error) {
	// Anthropic doesn't support embeddings
	// Return hash-based pseudo-embeddings for compatibility
	embeddings := make([]Embedding, len(texts))
	for i, text := range texts {
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

// CostPerToken returns the cost per token for Anthropic models
// Based on https://anthropic.com/pricing (approximate)
func (p *AnthropicProvider) CostPerToken(model string) float64 {
	costs := map[string]float64{
		"claude-3-5-sonnet-20241022": 0.000003, // $3/1M input, $15/1M output (average)
		"claude-3-5-sonnet":          0.000003,
		"claude-3-opus-20240229":     0.000015, // $15/1M input, $75/1M output
		"claude-3-opus":              0.000015,
		"claude-3-sonnet-20240229":   0.000003, // $3/1M input, $15/1M output
		"claude-3-sonnet":            0.000003,
		"claude-3-haiku-20240307":    0.0000008, // $0.80/1M input, $3.20/1M output
		"claude-3-haiku":             0.0000008,
	}

	if cost, ok := costs[model]; ok {
		return cost
	}

	// Default fallback
	return 0.000003
}

// IsAvailable checks if Anthropic API is available
func (p *AnthropicProvider) IsAvailable(ctx context.Context) bool {
	if p.config.APIKey == "" {
		return false
	}

	// Simple health check - try to list models
	req, err := http.NewRequestWithContext(ctx, "GET",
		p.config.Endpoint+"/v1/models", nil)
	if err != nil {
		return false
	}
	req.Header.Set("x-api-key", p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Anthropic doesn't have a public models endpoint, so we check with a simple request
	// Try a minimal message request
	testReq := map[string]interface{}{
		"model":      "claude-3-haiku-20240307",
		"max_tokens": 1,
		"messages":   []map[string]string{{"role": "user", "content": "hi"}},
	}
	jsonData, _ := json.Marshal(testReq)

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		p.config.Endpoint+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err = p.client.Do(httpReq)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 200 or 401 means API is reachable (401 = valid key but wrong permissions)
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized
}
