// Package memory provides the Mimir memory engine for ODIN
package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/odin-ai/odin/pkg/logger"
)

// Embedder generates embeddings for text
type Embedder interface {
	GenerateEmbedding(text string) ([]float32, error)
	Dimensions() int
	Name() string
}

// EmbeddingResult contains the embedding and metadata
type EmbeddingResult struct {
	Vector []float32
	Model  string
	Tokens int
}

// OllamaEmbedder generates real embeddings using Ollama
type OllamaEmbedder struct {
	endpoint string // "http://localhost:11434"
	model    string // "nomic-embed-text"
	timeout  time.Duration
}

// NewOllamaEmbedder creates a new Ollama embedder
func NewOllamaEmbedder(endpoint, model string) *OllamaEmbedder {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "nomic-embed-text"
	}
	return &OllamaEmbedder{
		endpoint: endpoint,
		model:    model,
		timeout:  30 * time.Second,
	}
}

// GenerateEmbedding creates a real vector embedding via Ollama API
func (e *OllamaEmbedder) GenerateEmbedding(text string) ([]float32, error) {
	reqBody := map[string]string{
		"model":  e.model,
		"prompt": text,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", e.endpoint+"/api/embeddings", bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: e.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return result.Embedding, nil
}

// Dimensions returns the embedding dimension count
func (e *OllamaEmbedder) Dimensions() int {
	// nomic-embed-text produces 768-dimensional vectors
	return 768
}

// Name returns the embedder name
func (e *OllamaEmbedder) Name() string {
	return "ollama:" + e.model
}

// OpenRouterEmbedder generates embeddings via OpenRouter API
type OpenRouterEmbedder struct {
	apiKey   string
	model    string
	endpoint string
	timeout  time.Duration
}

// NewOpenRouterEmbedder creates a new OpenRouter embedder
func NewOpenRouterEmbedder(apiKey, model string) *OpenRouterEmbedder {
	if model == "" {
		model = "openai/text-embedding-3-small"
	}
	return &OpenRouterEmbedder{
		apiKey:   apiKey,
		model:    model,
		endpoint: "https://openrouter.ai/api/v1/embeddings",
		timeout:  30 * time.Second,
	}
}

// GenerateEmbedding creates embedding via OpenRouter API
func (e *OpenRouterEmbedder) GenerateEmbedding(text string) ([]float32, error) {
	reqBody := map[string]string{
		"model": e.model,
		"input": text,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", e.endpoint, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	client := &http.Client{Timeout: e.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenRouter API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenRouter API returned status %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 || len(result.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return result.Data[0].Embedding, nil
}

// Dimensions returns the embedding dimension count (variable for OpenRouter)
func (e *OpenRouterEmbedder) Dimensions() int {
	return 1536 // text-embedding-3-small default
}

// Name returns the embedder name
func (e *OpenRouterEmbedder) Name() string {
	return "openrouter:" + e.model
}

// SimpleEmbeddingWrapper wraps SimpleVectorSearch to implement Embedder interface
type SimpleEmbeddingWrapper struct {
	*SimpleVectorSearch
}

// Dimensions returns the embedding dimension
func (w *SimpleEmbeddingWrapper) Dimensions() int {
	return 256
}

// Name returns the embedder name
func (w *SimpleEmbeddingWrapper) Name() string {
	return "simple:tfidf"
}

// DefaultEmbedder returns the default embedder with fallback chain:
// 1. OllamaEmbedder (nomic-embed-text) - local, 768 dims
// 2. OpenRouterEmbedder - if Ollama not available
// 3. SimpleVectorSearch - last resort (TF-IDF)
func DefaultEmbedder() Embedder {
	// Try Ollama first
	ollama := NewOllamaEmbedder("", "")

	// Test if Ollama is available by doing a quick probe
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", ollama.endpoint+"/api/tags", nil)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)

	if err == nil && resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		logger.Info("Using OllamaEmbedder", "model", ollama.model, "endpoint", ollama.endpoint)
		return ollama
	}

	if resp != nil {
		resp.Body.Close()
	}

	// Fall back to SimpleVectorSearch (TF-IDF placeholder)
	logger.Warn("Ollama not available, falling back to SimpleVectorSearch")
	return &SimpleEmbeddingWrapper{SimpleVectorSearch: NewSimpleVectorSearch()}
}

// EmbedderConfig holds configuration for the embedder
type EmbedderConfig struct {
	Type     string // "ollama", "openrouter", "simple"
	Endpoint string // For Ollama/OpenRouter
	Model    string // Model name
	APIKey   string // For OpenRouter
}

// NewEmbedderFromConfig creates an embedder from configuration
func NewEmbedderFromConfig(cfg *EmbedderConfig) Embedder {
	if cfg == nil {
		return DefaultEmbedder()
	}

	switch cfg.Type {
	case "ollama":
		embedder := NewOllamaEmbedder(cfg.Endpoint, cfg.Model)
		return embedder
	case "openrouter":
		if cfg.APIKey == "" {
			logger.Warn("OpenRouter API key not provided, using default embedder")
			return DefaultEmbedder()
		}
		return NewOpenRouterEmbedder(cfg.APIKey, cfg.Model)
	case "simple", "":
		return &SimpleEmbeddingWrapper{SimpleVectorSearch: NewSimpleVectorSearch()}
	default:
		logger.Warn("Unknown embedder type, using default", "type", cfg.Type)
		return DefaultEmbedder()
	}
}
