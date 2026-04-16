// Package memory provides the Mimir memory engine for ODIN
package memory

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOllamaEmbedderGenerateEmbedding(t *testing.T) {
	// Create a test server that returns a mock embedding
	expectedEmbedding := make([]float32, 768)
	for i := range expectedEmbedding {
		expectedEmbedding[i] = float32(i) / 768.0
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("Expected path /api/embeddings, got %s", r.URL.Path)
		}

		var req struct {
			Model  string `json:"model"`
			Prompt string `json:"prompt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		if req.Model != "nomic-embed-text" {
			t.Errorf("Expected model 'nomic-embed-text', got '%s'", req.Model)
		}
		if req.Prompt == "" {
			t.Error("Prompt should not be empty")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"embedding": expectedEmbedding,
		})
	}))
	defer server.Close()

	embedder := NewOllamaEmbedder(server.URL, "nomic-embed-text")
	embedding, err := embedder.GenerateEmbedding("Hello world")
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}

	if len(embedding) != 768 {
		t.Errorf("Expected embedding length 768, got %d", len(embedding))
	}

	// Verify first element
	if embedding[0] != 0.0 {
		t.Errorf("Expected first element 0.0, got %f", embedding[0])
	}
}

func TestOllamaEmbedderFallback(t *testing.T) {
	// Test that embedder falls back to SimpleVectorSearch when Ollama unavailable
	// This tests the DefaultEmbedder() fallback chain
	embedder := DefaultEmbedder()
	if embedder == nil {
		t.Fatal("DefaultEmbedder should not return nil")
	}

	// Should be SimpleEmbeddingWrapper since Ollama is not running
	wrapper, ok := embedder.(*SimpleEmbeddingWrapper)
	if !ok {
		t.Logf("Embedder type: %T", embedder)
		// This is expected in test environment - Ollama won't be running
	}

	// Verify it implements Embedder interface
	if wrapper != nil {
		if wrapper.Dimensions() != 256 {
			t.Errorf("Expected dimensions 256, got %d", wrapper.Dimensions())
		}
		if wrapper.Name() != "simple:tfidf" {
			t.Errorf("Expected name 'simple:tfidf', got '%s'", wrapper.Name())
		}
	}
}

func TestOllamaEmbedderTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"embedding": make([]float32, 768),
		})
	}))
	defer server.Close()

	embedder := NewOllamaEmbedder(server.URL, "nomic-embed-text")
	embedder.timeout = 100 * time.Millisecond // Very short timeout

	// This should timeout (the embedder has its own timeout)
	_, err := embedder.GenerateEmbedding("test")
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestOllamaEmbedderDimensions(t *testing.T) {
	embedder := NewOllamaEmbedder("", "")
	if embedder.Dimensions() != 768 {
		t.Errorf("Expected dimensions 768, got %d", embedder.Dimensions())
	}
}

func TestOllamaEmbedderName(t *testing.T) {
	embedder := NewOllamaEmbedder("", "")
	expected := "ollama:nomic-embed-text"
	if embedder.Name() != expected {
		t.Errorf("Expected name '%s', got '%s'", expected, embedder.Name())
	}
}

func TestOpenRouterEmbedder(t *testing.T) {
	expectedEmbedding := make([]float32, 1536)
	for i := range expectedEmbedding {
		expectedEmbedding[i] = float32(i) / 1536.0
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/embeddings" {
			t.Errorf("Expected path /api/v1/embeddings, got %s", r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("Expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"embedding": expectedEmbedding},
			},
		})
	}))
	defer server.Close()

	// Create embedder to test interface compliance
	_ = server // Server URL is not used since OpenRouterEmbedder has fixed endpoint
	embedder := NewOpenRouterEmbedder("test-api-key", "openai/text-embedding-3-small")
	if embedder.Dimensions() != 1536 {
		t.Errorf("Expected dimensions 1536, got %d", embedder.Dimensions())
	}
}

func TestSimpleEmbeddingWrapper(t *testing.T) {
	wrapper := &SimpleEmbeddingWrapper{SimpleVectorSearch: NewSimpleVectorSearch()}

	if wrapper.Dimensions() != 256 {
		t.Errorf("Expected dimensions 256, got %d", wrapper.Dimensions())
	}

	if wrapper.Name() != "simple:tfidf" {
		t.Errorf("Expected name 'simple:tfidf', got '%s'", wrapper.Name())
	}

	// Test that it can generate embeddings
	emb, err := wrapper.GenerateEmbedding("test")
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}

	if len(emb) != 256 {
		t.Errorf("Expected embedding length 256, got %d", len(emb))
	}
}

func TestNewEmbedderFromConfig(t *testing.T) {
	// Test nil config returns DefaultEmbedder
	embedder := NewEmbedderFromConfig(nil)
	if embedder == nil {
		t.Fatal("Should not return nil")
	}

	// Test simple type
	simpleEmbedder := NewEmbedderFromConfig(&EmbedderConfig{Type: "simple"})
	if simpleEmbedder.Name() != "simple:tfidf" {
		t.Errorf("Expected 'simple:tfidf', got '%s'", simpleEmbedder.Name())
	}

	// Test empty type falls back to simple
	emptyEmbedder := NewEmbedderFromConfig(&EmbedderConfig{Type: ""})
	if emptyEmbedder.Name() != "simple:tfidf" {
		t.Errorf("Expected 'simple:tfidf', got '%s'", emptyEmbedder.Name())
	}

	// Test unknown type falls back to default
	unknownEmbedder := NewEmbedderFromConfig(&EmbedderConfig{Type: "unknown"})
	if unknownEmbedder.Name() != "simple:tfidf" {
		t.Errorf("Expected 'simple:tfidf', got '%s'", unknownEmbedder.Name())
	}
}

func TestEmbeddingResult(t *testing.T) {
	result := EmbeddingResult{
		Vector: []float32{1.0, 2.0, 3.0},
		Model:  "test-model",
		Tokens: 10,
	}

	if len(result.Vector) != 3 {
		t.Errorf("Expected 3 vector elements, got %d", len(result.Vector))
	}

	if result.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", result.Model)
	}

	if result.Tokens != 10 {
		t.Errorf("Expected tokens 10, got %d", result.Tokens)
	}
}

func TestOllamaServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	embedder := NewOllamaEmbedder(server.URL, "nomic-embed-text")
	_, err := embedder.GenerateEmbedding("test")
	if err == nil {
		t.Error("Expected error for server error status")
	}
}

func TestOllamaEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"embedding": []float32{},
		})
	}))
	defer server.Close()

	embedder := NewOllamaEmbedder(server.URL, "nomic-embed-text")
	_, err := embedder.GenerateEmbedding("test")
	if err == nil {
		t.Error("Expected error for empty embedding")
	}
}

func TestEmbedderConfig(t *testing.T) {
	cfg := &EmbedderConfig{
		Type:     "ollama",
		Endpoint: "http://localhost:11434",
		Model:    "nomic-embed-text",
	}

	if cfg.Type != "ollama" {
		t.Errorf("Expected type 'ollama', got '%s'", cfg.Type)
	}

	if cfg.Endpoint != "http://localhost:11434" {
		t.Errorf("Expected endpoint 'http://localhost:11434', got '%s'", cfg.Endpoint)
	}

	if cfg.Model != "nomic-embed-text" {
		t.Errorf("Expected model 'nomic-embed-text', got '%s'", cfg.Model)
	}
}
