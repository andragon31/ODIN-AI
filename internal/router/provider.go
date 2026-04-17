// Package router provides model routing with fallback chain for ODIN
package router

import (
	"context"
	"time"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GenerateRequest is a request for text generation
type GenerateRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Usage represents token usage statistics
type Usage struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	TotalTokens  int     `json:"total_tokens"`
	Cost         float64 `json:"cost"`
}

// GenerateResponse is a response from text generation
type GenerateResponse struct {
	Content string `json:"content"`
	Usage   Usage  `json:"usage"`
	Model   string `json:"model"`
}

// Embedding represents a vector embedding
type Embedding struct {
	Vector  []float64 `json:"vector"`
	Content string    `json:"content,omitempty"`
}

// Provider is the interface for model providers
type Provider interface {
	// Name returns the provider name
	Name() string
	// Supports checks if the provider supports a given model
	Supports(model string) bool
	// Generate generates text from the model
	Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
	// Embed generates embeddings for texts
	Embed(ctx context.Context, texts []string) ([]Embedding, error)
	// CostPerToken returns the cost per token for a model (0 for free providers)
	CostPerToken(model string) float64
	// IsAvailable checks if the provider is currently available
	IsAvailable(ctx context.Context) bool
}

// Timeout is the default timeout for provider requests
const Timeout = 30 * time.Second
