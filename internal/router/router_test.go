package router

import (
	"context"
	"testing"
	"time"
)

// MockProvider is a mock provider for testing
type MockProvider struct {
	name       string
	supported  bool
	available  bool
	shouldFail bool
	latency    time.Duration
}

func (m *MockProvider) Name() string                         { return m.name }
func (m *MockProvider) Supports(model string) bool           { return m.supported }
func (m *MockProvider) IsAvailable(ctx context.Context) bool { return m.available }
func (m *MockProvider) CostPerToken(model string) float64    { return 0 }

func (m *MockProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	if m.shouldFail {
		return nil, context.DeadlineExceeded
	}
	time.Sleep(m.latency)
	return &GenerateResponse{
		Content: "mock response",
		Usage: Usage{
			InputTokens:  10,
			OutputTokens: 20,
			TotalTokens:  30,
			Cost:         0,
		},
		Model: req.Model,
	}, nil
}

func (m *MockProvider) Embed(ctx context.Context, texts []string) ([]Embedding, error) {
	embeddings := make([]Embedding, len(texts))
	for i := range texts {
		embeddings[i] = Embedding{
			Vector:  []float64{0.1, 0.2, 0.3},
			Content: texts[i],
		}
	}
	return embeddings, nil
}

func TestNewRouter(t *testing.T) {
	providers := []Provider{
		&MockProvider{name: "provider1", supported: true, available: true},
		&MockProvider{name: "provider2", supported: true, available: true},
	}

	r, err := NewRouter(providers, "provider1")
	if err != nil {
		t.Fatalf("NewRouter failed: %v", err)
	}

	if r == nil {
		t.Fatal("NewRouter returned nil")
	}

	defaultProvider := r.GetDefaultProvider()
	if defaultProvider.Name() != "provider1" {
		t.Errorf("Expected default provider 'provider1', got '%s'", defaultProvider.Name())
	}
}

func TestRouterGenerateSuccess(t *testing.T) {
	providers := []Provider{
		&MockProvider{name: "good", supported: true, available: true, latency: 1 * time.Millisecond},
	}

	r, err := NewRouter(providers, "good")
	if err != nil {
		t.Fatalf("NewRouter failed: %v", err)
	}

	ctx := context.Background()
	req := GenerateRequest{
		Model:       "test-model",
		Messages:    []Message{{Role: "user", Content: "hello"}},
		MaxTokens:   100,
		Temperature: 0.7,
	}

	resp, err := r.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Content != "mock response" {
		t.Errorf("Expected 'mock response', got '%s'", resp.Content)
	}

	if resp.Usage.TotalTokens != 30 {
		t.Errorf("Expected 30 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestRouterGenerateFallback(t *testing.T) {
	providers := []Provider{
		&MockProvider{name: "failing", supported: true, available: true, shouldFail: true},
		&MockProvider{name: "good", supported: true, available: true, latency: 1 * time.Millisecond},
	}

	r, err := NewRouter(providers, "failing")
	if err != nil {
		t.Fatalf("NewRouter failed: %v", err)
	}

	ctx := context.Background()
	req := GenerateRequest{
		Model:    "test-model",
		Messages: []Message{{Role: "user", Content: "hello"}},
	}

	resp, err := r.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Content != "mock response" {
		t.Errorf("Expected fallback to 'good' provider, got '%s'", resp.Content)
	}
}

func TestRouterGenerateAllFail(t *testing.T) {
	providers := []Provider{
		&MockProvider{name: "failing1", supported: true, available: true, shouldFail: true},
		&MockProvider{name: "failing2", supported: true, available: true, shouldFail: true},
	}

	r, err := NewRouter(providers, "failing1")
	if err != nil {
		t.Fatalf("NewRouter failed: %v", err)
	}

	ctx := context.Background()
	req := GenerateRequest{
		Model:    "test-model",
		Messages: []Message{{Role: "user", Content: "hello"}},
	}

	_, err = r.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error when all providers fail")
	}
}

func TestRouterFallbackChain(t *testing.T) {
	providers := []Provider{
		&MockProvider{name: "first", supported: true, available: true, shouldFail: true},
		&MockProvider{name: "second", supported: true, available: true, shouldFail: true},
		&MockProvider{name: "third", supported: true, available: true, latency: 1 * time.Millisecond},
	}

	r, err := NewRouter(providers, "first")
	if err != nil {
		t.Fatalf("NewRouter failed: %v", err)
	}

	// Set fallback chain: first -> second -> third
	r.SetFallbackChain([]string{"first", "second", "third"})

	ctx := context.Background()
	req := GenerateRequest{
		Model:    "test-model",
		Messages: []Message{{Role: "user", Content: "hello"}},
	}

	resp, err := r.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Content != "mock response" {
		t.Errorf("Expected fallback to 'third' provider, got '%s'", resp.Content)
	}
}

func TestRouterUnsupportedModel(t *testing.T) {
	providers := []Provider{
		&MockProvider{name: "provider", supported: false, available: true},
	}

	r, err := NewRouter(providers, "provider")
	if err != nil {
		t.Fatalf("NewRouter failed: %v", err)
	}

	ctx := context.Background()
	req := GenerateRequest{
		Model:    "unsupported-model",
		Messages: []Message{{Role: "user", Content: "hello"}},
	}

	_, err = r.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error for unsupported model")
	}
}

func TestRouterEmbed(t *testing.T) {
	providers := []Provider{
		&MockProvider{name: "provider", supported: true, available: true},
	}

	r, err := NewRouter(providers, "provider")
	if err != nil {
		t.Fatalf("NewRouter failed: %v", err)
	}

	ctx := context.Background()
	texts := []string{"hello", "world"}

	embeddings, err := r.Embed(ctx, texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(embeddings) != 2 {
		t.Errorf("Expected 2 embeddings, got %d", len(embeddings))
	}
}

func TestRouterCheckHealth(t *testing.T) {
	providers := []Provider{
		&MockProvider{name: "available", supported: true, available: true},
		&MockProvider{name: "unavailable", supported: true, available: false},
	}

	r, err := NewRouter(providers, "available")
	if err != nil {
		t.Fatalf("NewRouter failed: %v", err)
	}

	ctx := context.Background()
	health := r.CheckHealth(ctx)

	if !health["available"] {
		t.Error("Expected 'available' to be healthy")
	}

	if health["unavailable"] {
		t.Error("Expected 'unavailable' to be unhealthy")
	}
}

func TestRouterGetProvider(t *testing.T) {
	providers := []Provider{
		&MockProvider{name: "provider1", supported: true, available: true},
		&MockProvider{name: "provider2", supported: true, available: true},
	}

	r, err := NewRouter(providers, "provider1")
	if err != nil {
		t.Fatalf("NewRouter failed: %v", err)
	}

	p := r.GetProvider("provider2")
	if p == nil {
		t.Error("GetProvider returned nil for existing provider")
	}

	p = r.GetProvider("nonexistent")
	if p != nil {
		t.Error("GetProvider should return nil for nonexistent provider")
	}
}

func TestMetricsRecordSuccess(t *testing.T) {
	m := NewMetrics()
	m.RecordSuccess("provider1", 100*time.Millisecond, 50)

	metrics := m.GetProviderMetrics("provider1")
	if metrics == nil {
		t.Fatal("GetProviderMetrics returned nil")
	}

	if metrics.RequestCount != 1 {
		t.Errorf("Expected 1 request, got %d", metrics.RequestCount)
	}

	if metrics.SuccessCount != 1 {
		t.Errorf("Expected 1 success, got %d", metrics.SuccessCount)
	}

	if metrics.TotalTokens != 50 {
		t.Errorf("Expected 50 tokens, got %d", metrics.TotalTokens)
	}
}

func TestMetricsRecordError(t *testing.T) {
	m := NewMetrics()
	m.RecordError("provider1")

	metrics := m.GetProviderMetrics("provider1")
	if metrics == nil {
		t.Fatal("GetProviderMetrics returned nil")
	}

	if metrics.RequestCount != 1 {
		t.Errorf("Expected 1 request, got %d", metrics.RequestCount)
	}

	if metrics.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", metrics.ErrorCount)
	}
}

func TestMetricsSuccessRate(t *testing.T) {
	m := NewMetrics()
	m.RecordSuccess("provider1", 100*time.Millisecond, 50)
	m.RecordSuccess("provider1", 100*time.Millisecond, 50)
	m.RecordError("provider1")
	m.RecordError("provider1")

	rate := m.SuccessRate("provider1")
	if rate != 50 {
		t.Errorf("Expected 50%% success rate, got %.2f%%", rate)
	}
}

func TestMetricsAverageLatency(t *testing.T) {
	m := NewMetrics()
	m.RecordSuccess("provider1", 100*time.Millisecond, 50)
	m.RecordSuccess("provider1", 200*time.Millisecond, 50)

	avg := m.AverageLatency("provider1")
	expected := 150 * time.Millisecond

	if avg != expected {
		t.Errorf("Expected average latency %v, got %v", expected, avg)
	}
}

func TestMetricsReset(t *testing.T) {
	m := NewMetrics()
	m.RecordSuccess("provider1", 100*time.Millisecond, 50)

	m.Reset()

	metrics := m.GetProviderMetrics("provider1")
	if metrics != nil {
		t.Error("Expected nil metrics after reset")
	}
}
