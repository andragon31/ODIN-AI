// Package runeforge provides rune generation via ODIN's local Router
// This is ODIN's equivalent of AgentBuilder - it generates new skills using
// the local Router (Ollama/OpenRouter/Anthropic) instead of external AI CLIs
package runeforge

import (
	"context"
	"fmt"
	"time"

	"github.com/odin-ai/odin/internal/router"
	"github.com/odin-ai/odin/internal/skills"
	"github.com/odin-ai/odin/pkg/logger"
)

// RuneForge generates new runes using the local Router
type RuneForge struct {
	router *router.Router
	parser *Parser
}

// ForgeRequest contains the parameters for rune generation
type ForgeRequest struct {
	Name        string   // Name of the rune to generate
	Description string   // Description of what the rune does
	Tags        []string // Tags for categorization
	Model       string   // Model to use, e.g., "ollama:deepseek-coder"
}

// ForgeResult contains the result of rune generation
type ForgeResult struct {
	Rune   *skills.Rune // The generated rune
	Valid  bool         // Whether the rune passed validation
	Errors []string     // Validation errors (if any)
}

// NewRuneForge creates a new RuneForge instance
func NewRuneForge(r *router.Router) *RuneForge {
	return &RuneForge{
		router: r,
		parser: NewParser(),
	}
}

// Generate creates a new rune using the local Router
func (f *RuneForge) Generate(ctx context.Context, req ForgeRequest) (*ForgeResult, error) {
	// Build the generation prompt
	prompt := f.buildPrompt(req)

	// Use the model from request or default
	model := req.Model
	if model == "" {
		model = "ollama:deepseek-coder"
	}

	// Generate using local Router (works offline with Ollama)
	logger.Info("Generating rune via Router", "model", model, "name", req.Name)

	response, err := f.router.Generate(ctx, router.GenerateRequest{
		Model: model,
		Messages: []router.Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   2048,
		Temperature: 0.7,
	})

	if err != nil {
		return nil, fmt.Errorf("Router generation failed: %w", err)
	}

	// Parse the model output into a Rune struct
	rune, err := f.parser.ParseRune(response.Content)
	if err != nil {
		return &ForgeResult{
			Rune:   nil,
			Valid:  false,
			Errors: []string{fmt.Sprintf("parse error: %v", err)},
		}, nil
	}

	// Set metadata
	rune.Name = req.Name
	if req.Description != "" {
		rune.Description = req.Description
	}
	if len(req.Tags) > 0 {
		rune.Tags = req.Tags
	}

	// Validate the generated rune
	result := f.ValidateRune(rune)

	return result, nil
}

// GenerateFromExample creates a rune based on an existing example
func (f *RuneForge) GenerateFromExample(ctx context.Context, examplePath string, adaptFor string) (*ForgeResult, error) {
	// Read the example rune
	exampleRune, validationResult, err := skills.ValidateFile(examplePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read example: %w", err)
	}

	if !validationResult.Valid {
		return nil, fmt.Errorf("example is not a valid rune: %v", validationResult.Errors)
	}

	// Build adaptation prompt
	prompt := f.buildAdaptPrompt(exampleRune, adaptFor)

	model := "ollama:deepseek-coder"

	response, err := f.router.Generate(ctx, router.GenerateRequest{
		Model: model,
		Messages: []router.Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   2048,
		Temperature: 0.7,
	})

	if err != nil {
		return nil, fmt.Errorf("Router generation failed: %w", err)
	}

	// Parse the adapted rune
	rune, err := f.parser.ParseRune(response.Content)
	if err != nil {
		return &ForgeResult{
			Rune:   nil,
			Valid:  false,
			Errors: []string{fmt.Sprintf("parse error: %v", err)},
		}, nil
	}

	// Set source as adapted
	if rune.Description == "" {
		rune.Description = fmt.Sprintf("Adapted from %s for %s", exampleRune.Name, adaptFor)
	}

	result := f.ValidateRune(rune)
	return result, nil
}

// ValidateRune validates a rune against the schema
func (f *RuneForge) ValidateRune(r *skills.Rune) *ForgeResult {
	result := &ForgeResult{
		Rune:   r,
		Valid:  true,
		Errors: []string{},
	}

	// Run schema validation
	validationResult := skills.ValidateSkill(r)

	if !validationResult.Valid {
		result.Valid = false
		result.Errors = append(result.Errors, validationResult.Errors...)
	}

	// Add warnings but don't fail
	if len(validationResult.Warns) > 0 {
		logger.Warn("Rune validation warnings", "warnings", validationResult.Warns)
	}

	return result
}

// GenerateAsync creates a rune asynchronously
func (f *RuneForge) GenerateAsync(ctx context.Context, req ForgeRequest) (<-chan *ForgeResult, <-chan error) {
	resultCh := make(chan *ForgeResult, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := f.Generate(ctx, req)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	return resultCh, errCh
}

// GenerateWithTimeout creates a rune with a timeout
func (f *RuneForge) GenerateWithTimeout(ctx context.Context, req ForgeRequest, timeout time.Duration) (*ForgeResult, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultCh, errCh := f.GenerateAsync(ctx, req)

	select {
	case result := <-resultCh:
		return result, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("generation timed out after %v", timeout)
	}
}
