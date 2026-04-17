package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/odin-ai/odin/internal/config"
)

// OpenAIProvider is a provider for OpenAI-compatible models
type OpenAIProvider struct {
	config config.OpenAIConfig
	client *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(cfg config.OpenAIConfig) *OpenAIProvider {
	if cfg.Endpoint == "" {
		cfg.Endpoint = config.DefaultOpenAIEndpoint
	}

	return &OpenAIProvider{
		config: cfg,
		client: &http.Client{
			Timeout: Timeout,
		},
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Supports checks if the provider supports a given model
func (p *OpenAIProvider) Supports(model string) bool {
	// Standard OpenAI models
	supportedModels := map[string]bool{
		"gpt-4o":            true,
		"gpt-4-turbo":       true,
		"gpt-4":             true,
		"gpt-3.5-turbo":     true,
		"text-embedding-3-small": true,
	}

	if supportedModels[model] {
		return true
	}

	// Also support custom models if provider is configured
	return p.config.APIKey != ""
}

// Generate generates text from the model
func (p *OpenAIProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	if req.MaxTokens == 0 {
		req.MaxTokens = 2048
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	openAIReq := map[string]interface{}{
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

	jsonData, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := p.config.Endpoint
	if endpoint[len(endpoint)-1] == '/' {
		endpoint = endpoint[:len(endpoint)-1]
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		endpoint+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("openai error: %s (type %s)", errResp.Error.Message, errResp.Error.Type)
	}

	var openAIResp struct {
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

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openai")
	}

	return &GenerateResponse{
		Content: openAIResp.Choices[0].Message.Content,
		Usage: Usage{
			InputTokens:  openAIResp.Usage.PromptTokens,
			OutputTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:  openAIResp.Usage.TotalTokens,
			Cost:         p.CostPerToken(req.Model) * float64(openAIResp.Usage.TotalTokens),
		},
		Model: req.Model,
	}, nil
}

// Embed generates embeddings
func (p *OpenAIProvider) Embed(ctx context.Context, texts []string) ([]Embedding, error) {
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("openai provider not configured with API key")
	}

	req := map[string]interface{}{
		"input": texts,
		"model": "text-embedding-3-small",
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	endpoint := p.config.Endpoint
	if endpoint[len(endpoint)-1] == '/' {
		endpoint = endpoint[:len(endpoint)-1]
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		endpoint+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai embeddings error: status %d", resp.StatusCode)
	}

	var openAIResp struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	embeddings := make([]Embedding, len(openAIResp.Data))
	for i, d := range openAIResp.Data {
		embeddings[i] = Embedding{
			Vector:  d.Embedding,
			Content: texts[i],
		}
	}

	return embeddings, nil
}

// CostPerToken returns the cost per token
func (p *OpenAIProvider) CostPerToken(model string) float64 {
	costs := map[string]float64{
		"gpt-4o":        0.000005, // $5/1M
		"gpt-4-turbo":   0.00001,  // $10/1M
		"gpt-4":         0.00003,  // $30/1M
		"gpt-3.5-turbo": 0.0000005, // $0.50/1M
	}

	if cost, ok := costs[model]; ok {
		return cost
	}
	return 0.000001
}

// IsAvailable checks if OpenAI API is available
func (p *OpenAIProvider) IsAvailable(ctx context.Context) bool {
	return p.config.Enabled && p.config.APIKey != ""
}
