package router

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/odin-ai/odin/pkg/logger"
)

// Router manages model providers with fallback chain
type Router struct {
	providers    []Provider
	defaultIdx   int
	metrics      *Metrics
	mu           sync.RWMutex
	fallbackList []string
}

// NewRouter creates a new router with the given providers
func NewRouter(providers []Provider, defaultProvider string) (*Router, error) {
	r := &Router{
		providers:    providers,
		metrics:      NewMetrics(),
		fallbackList: []string{},
	}

	// Set default provider index
	for i, p := range providers {
		if p.Name() == defaultProvider {
			r.defaultIdx = i
			break
		}
	}

	return r, nil
}

// SetFallbackChain sets the fallback chain order
func (r *Router) SetFallbackChain(providers []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallbackList = providers
}

// Generate sends a request to the best available provider
func (r *Router) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	r.mu.RLock()
	providers := make([]Provider, len(r.providers))
	copy(providers, r.providers)
	fallbackList := make([]string, len(r.fallbackList))
	copy(fallbackList, r.fallbackList)
	defaultIdx := r.defaultIdx
	r.mu.RUnlock()

	// If fallback chain is set, use it; otherwise try providers in order
	if len(fallbackList) > 0 {
		return r.generateWithFallbackChain(ctx, req, fallbackList, providers)
	}

	// Try default provider first
	startTime := time.Now()
	provider := providers[defaultIdx]

	if provider.Supports(req.Model) && provider.IsAvailable(ctx) {
		resp, err := provider.Generate(ctx, req)
		if err == nil {
			r.metrics.RecordSuccess(provider.Name(), time.Since(startTime), resp.Usage.TotalTokens)
			return resp, nil
		}
		r.metrics.RecordError(provider.Name())
		logger.Warn("Provider failed", "provider", provider.Name(), "error", err)
	}

	// Try other providers in order
	for i, p := range providers {
		if i == defaultIdx {
			continue // already tried
		}

		if !p.Supports(req.Model) {
			continue
		}

		startTime = time.Now()
		if !p.IsAvailable(ctx) {
			r.metrics.RecordError(p.Name())
			continue
		}

		resp, err := p.Generate(ctx, req)
		if err == nil {
			r.metrics.RecordSuccess(p.Name(), time.Since(startTime), resp.Usage.TotalTokens)
			return resp, nil
		}
		r.metrics.RecordError(p.Name())
		logger.Warn("Provider failed", "provider", p.Name(), "error", err)
	}

	return nil, fmt.Errorf("all providers failed for model %s", req.Model)
}

// generateWithFallbackChain tries providers in the fallback chain order
func (r *Router) generateWithFallbackChain(ctx context.Context, req GenerateRequest, chain []string, providers []Provider) (*GenerateResponse, error) {
	providerMap := make(map[string]Provider)
	for _, p := range providers {
		providerMap[p.Name()] = p
	}

	for _, name := range chain {
		p, ok := providerMap[name]
		if !ok {
			continue
		}

		if !p.Supports(req.Model) {
			logger.Debug("Provider doesn't support model", "provider", name, "model", req.Model)
			continue
		}

		startTime := time.Now()
		if !p.IsAvailable(ctx) {
			r.metrics.RecordError(p.Name())
			logger.Debug("Provider not available", "provider", name)
			continue
		}

		resp, err := p.Generate(ctx, req)
		if err == nil {
			r.metrics.RecordSuccess(p.Name(), time.Since(startTime), resp.Usage.TotalTokens)
			return resp, nil
		}
		r.metrics.RecordError(p.Name())
		logger.Warn("Provider failed in fallback chain", "provider", name, "error", err)
	}

	return nil, fmt.Errorf("all providers in fallback chain failed for model %s", req.Model)
}

// Embed sends an embedding request to the best available provider
func (r *Router) Embed(ctx context.Context, texts []string) ([]Embedding, error) {
	r.mu.RLock()
	providers := make([]Provider, len(r.providers))
	copy(providers, r.providers)
	defaultIdx := r.defaultIdx
	r.mu.RUnlock()

	// Try default provider first
	provider := providers[defaultIdx]
	if provider.IsAvailable(ctx) {
		embeddings, err := provider.Embed(ctx, texts)
		if err == nil {
			return embeddings, nil
		}
		logger.Warn("Provider embedding failed", "provider", provider.Name(), "error", err)
	}

	// Try other providers
	for i, p := range providers {
		if i == defaultIdx {
			continue
		}
		if !p.IsAvailable(ctx) {
			continue
		}
		embeddings, err := p.Embed(ctx, texts)
		if err == nil {
			return embeddings, nil
		}
		logger.Warn("Provider embedding failed", "provider", p.Name(), "error", err)
	}

	return nil, fmt.Errorf("all providers failed for embedding")
}

// GetProvider returns a provider by name
func (r *Router) GetProvider(name string) Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.providers {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// GetDefaultProvider returns the default provider
func (r *Router) GetDefaultProvider() Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[r.defaultIdx]
}

// ListProviders returns all providers
func (r *Router) ListProviders() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	providers := make([]Provider, len(r.providers))
	copy(providers, r.providers)
	return providers
}

// GetMetrics returns the router metrics
func (r *Router) GetMetrics() *Metrics {
	return r.metrics
}

// CheckHealth checks the health of all providers
func (r *Router) CheckHealth(ctx context.Context) map[string]bool {
	r.mu.RLock()
	providers := make([]Provider, len(r.providers))
	copy(providers, r.providers)
	r.mu.RUnlock()

	results := make(map[string]bool)
	for _, p := range providers {
		results[p.Name()] = p.IsAvailable(ctx)
	}
	return results
}
