package runeforge

import (
	"context"

	"github.com/odin-ai/odin/internal/router"
	"github.com/odin-ai/odin/internal/skills"
)

// Engine is the interface for rune generation engines
type Engine interface {
	// Generate generates a rune from a prompt
	Generate(ctx context.Context, prompt string, model string) (string, error)
	// Parse parses model output into a rune
	Parse(content string) (*skills.Rune, error)
}

// RouterEngine is an Engine implementation using the Router
type RouterEngine struct {
	router *router.Router
}

// NewRouterEngine creates a new Router-based engine
func NewRouterEngine(r *router.Router) *RouterEngine {
	return &RouterEngine{router: r}
}

// Generate generates text using the router
func (e *RouterEngine) Generate(ctx context.Context, prompt string, model string) (string, error) {
	if model == "" {
		model = "ollama:deepseek-coder"
	}

	response, err := e.router.Generate(ctx, router.GenerateRequest{
		Model: model,
		Messages: []router.Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   2048,
		Temperature: 0.7,
	})

	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// Parse parses text into a rune using the default parser
func (e *RouterEngine) Parse(content string) (*skills.Rune, error) {
	parser := NewParser()
	return parser.ParseRune(content)
}

// GenerationStrategy defines how to generate content
type GenerationStrategy string

const (
	// StrategyDirect generates runes directly from description
	StrategyDirect GenerationStrategy = "direct"
	// StrategyExample generates based on an example
	StrategyExample GenerationStrategy = "example"
	// StrategyIterative generates with refinement
	StrategyIterative GenerationStrategy = "iterative"
)

// EngineConfig contains configuration for the engine
type EngineConfig struct {
	Strategy      GenerationStrategy
	Model         string
	MaxTokens     int
	Temperature   float64
	MaxIterations int
}

// DefaultEngineConfig returns the default engine configuration
func DefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		Strategy:      StrategyDirect,
		Model:         "ollama:deepseek-coder",
		MaxTokens:     2048,
		Temperature:   0.7,
		MaxIterations: 3,
	}
}
