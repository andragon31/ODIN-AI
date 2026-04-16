package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OllamaProvider is a provider for local Ollama models
type OllamaProvider struct {
	config OllamaConfig
	client *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config OllamaConfig) *OllamaProvider {
	if config.Endpoint == "" {
		config.Endpoint = DefaultOllamaEndpoint
	}

	return &OllamaProvider{
		config: config,
		client: &http.Client{
			Timeout: Timeout,
		},
	}
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// Supports checks if the provider supports a given model
func (p *OllamaProvider) Supports(model string) bool {
	// Ollama supports any model that has been pulled
	// We assume all models are supported if Ollama is running
	return true
}

// Generate generates text from the model
func (p *OllamaProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	if req.MaxTokens == 0 {
		req.MaxTokens = 2048
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	// Convert messages to Ollama format
	ollamaReq := map[string]interface{}{
		"model":  req.Model,
		"stream": false,
		"options": map[string]interface{}{
			"num_predict": req.MaxTokens,
			"temperature": req.Temperature,
		},
	}

	// Build prompt from messages
	var prompt string
	for _, msg := range req.Messages {
		prompt += fmt.Sprintf("<|%s|> %s </|>", msg.Role, msg.Content)
	}
	ollamaReq["prompt"] = prompt

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		p.config.Endpoint+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var ollamaResp struct {
		Response        string `json:"response"`
		PromptEvalCount int    `json:"prompt_eval_count"`
		EvalCount       int    `json:"eval_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &GenerateResponse{
		Content: ollamaResp.Response,
		Usage: Usage{
			InputTokens:  ollamaResp.PromptEvalCount,
			OutputTokens: ollamaResp.EvalCount,
			TotalTokens:  ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
			Cost:         0, // Ollama is free
		},
		Model: req.Model,
	}, nil
}

// Embed generates embeddings for texts
func (p *OllamaProvider) Embed(ctx context.Context, texts []string) ([]Embedding, error) {
	embeddings := make([]Embedding, len(texts))

	for i, text := range texts {
		ollamaReq := map[string]interface{}{
			"model":  "nomic-embed-text",
			"prompt": text,
		}

		jsonData, err := json.Marshal(ollamaReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal embed request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST",
			p.config.Endpoint+"/api/embeddings", bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create embed request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to send embed request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("ollama embeddings returned status %d", resp.StatusCode)
		}

		var embedResp struct {
			Embedding []float64 `json:"embedding"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode embed response: %w", err)
		}
		resp.Body.Close()

		embeddings[i] = Embedding{
			Vector:  embedResp.Embedding,
			Content: text,
		}
	}

	return embeddings, nil
}

// CostPerToken returns the cost per token (0 for Ollama)
func (p *OllamaProvider) CostPerToken(model string) float64 {
	return 0
}

// IsAvailable checks if Ollama is running and available
func (p *OllamaProvider) IsAvailable(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.config.Endpoint+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
