// Package memory provides the Mimir memory engine for ODIN
// Mimir is the Norse god of wisdom and knowledge, representing
// the memory/knowledge graph functionality of ODIN AI
package memory

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/odin-ai/odin/pkg/logger"
)

// Memory represents a single memory entry in Mimir
type Memory struct {
	ID         string                 `json:"id"`
	Content    string                 `json:"content"`
	Embedding  []float32              `json:"embedding,omitempty"`
	Project    string                 `json:"project,omitempty"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Encrypted  bool                   `json:"encrypted"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	AccessedAt time.Time              `json:"accessed_at"`
}

// MemoryEdge represents a relationship between memories in the knowledge graph
type MemoryEdge struct {
	FromID   string  `json:"from_id"`
	ToID     string  `json:"to_id"`
	Relation string  `json:"relation"` // e.g., "inspired_by", "depends_on", "contradicts"
	Weight   float64 `json:"weight"`
}

// SearchResult represents a search result with similarity score
type SearchResult struct {
	Memory   *Memory `json:"memory"`
	Score    float64 `json:"score"`
	Distance float64 `json:"distance"`
}

// Config holds Mimir configuration
type Config struct {
	DBPath        string
	EncryptionKey string
	KeepTags      []string
	PruneInterval time.Duration
	Remote        string
	VSSEnabled    bool
}

// DefaultConfig returns the default Mimir configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		DBPath:        filepath.Join(homeDir, ".odin", "memory.db"),
		EncryptionKey: "",
		KeepTags:      []string{"arch", "spec", "security"},
		PruneInterval: 24 * time.Hour,
		Remote:        "",
		VSSEnabled:    true,
	}
}

// Store represents the Mimir memory store
type Store struct {
	db     *DB
	config *Config
	vss    VectorSearcher
	enc    *Encryptor
	sync   *Syncer
}

// NewStore creates a new Mimir store
func NewStore(cfg *Config) (*Store, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Ensure directory exists
	dir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create memory directory: %w", err)
	}

	db, err := NewDB(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{
		db:     db,
		config: cfg,
		vss:    NewSimpleVectorSearch(),
		enc:    NewEncryptor(),
		sync:   NewSyncer(db),
	}

	// Try to initialize vector search
	if cfg.VSSEnabled {
		if err := store.initVectorSearch(); err != nil {
			logger.Warn("Vector search initialization failed, using FTS5 fallback",
				"error", err)
			store.vss = nil // Will use FTS5 fallback
		}
	}

	return store, nil
}

// initVectorSearch initializes the vector search engine
func (s *Store) initVectorSearch() error {
	// Create vector search table if not exists
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS memory_vectors (
			memory_id TEXT PRIMARY KEY,
			embedding BLOB NOT NULL,
			FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create vectors table: %w", err)
	}
	return nil
}

// Close closes the store and underlying database
func (s *Store) Close() error {
	return s.db.Close()
}

// Store saves a memory to the store
func (s *Store) Store(m *Memory) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	m.AccessedAt = now

	// Generate embedding if not present and vector search is enabled
	if s.vss != nil && len(m.Embedding) == 0 {
		embedding, err := s.vss.GenerateEmbedding(m.Content)
		if err != nil {
			logger.Warn("Failed to generate embedding", "error", err)
		} else {
			m.Embedding = embedding
			// Store vector
			if err := s.db.StoreVector(m.ID, embedding); err != nil {
				logger.Warn("Failed to store vector", "error", err)
			}
		}
	}

	// Serialize metadata
	var metadataJSON []byte
	if m.Metadata != nil {
		metadataJSON, _ = json.Marshal(m.Metadata)
	}

	// Insert or update memory
	err := s.db.UpsertMemory(m, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to store memory: %w", err)
	}

	logger.Debug("Memory stored", "id", m.ID, "tags", m.Tags)
	return nil
}

// Recall retrieves a memory by ID
func (s *Store) Recall(id string) (*Memory, error) {
	m, metadataJSON, err := s.db.GetMemory(id)
	if err != nil {
		return nil, fmt.Errorf("failed to recall memory: %w", err)
	}

	if m != nil && metadataJSON != nil {
		json.Unmarshal(metadataJSON, &m.Metadata)
	}

	// Update accessed time
	s.db.UpdateAccessedAt(id)

	return m, nil
}

// Search performs semantic search on memories
func (s *Store) Search(query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	// If vector search is available, use it
	if s.vss != nil {
		embedding, err := s.vss.GenerateEmbedding(query)
		if err != nil {
			return nil, fmt.Errorf("failed to generate query embedding: %w", err)
		}

		results, err := s.db.VectorSearch(embedding, limit)
		if err != nil {
			logger.Warn("Vector search failed, falling back to FTS5",
				"error", err)
			return s.fts5Search(query, limit)
		}

		// Enrich results with full memory data
		searchResults := make([]SearchResult, 0, len(results))
		for _, r := range results {
			m, _, err := s.db.GetMemory(r.MemoryID)
			if err != nil {
				continue
			}
			if m != nil {
				searchResults = append(searchResults, SearchResult{
					Memory:   m,
					Score:    r.Score,
					Distance: r.Distance,
				})
			}
		}

		return searchResults, nil
	}

	// Fallback to FTS5
	return s.fts5Search(query, limit)
}

// fts5Search performs FTS5 full-text search as fallback
func (s *Store) fts5Search(query string, limit int) ([]SearchResult, error) {
	memories, err := s.db.FTSSearch(query, limit)
	if err != nil {
		return nil, fmt.Errorf("fts5 search failed: %w", err)
	}

	results := make([]SearchResult, 0, len(memories))
	for _, m := range memories {
		results = append(results, SearchResult{
			Memory:   m,
			Score:    1.0, // FTS5 doesn't provide scores
			Distance: 0.0,
		})
	}

	return results, nil
}

// ListTags returns all unique tags in the store
func (s *Store) ListTags() ([]string, error) {
	return s.db.ListTags()
}

// AddTag adds a tag to a memory
func (s *Store) AddTag(memoryID, tag string) error {
	return s.db.AddTag(memoryID, tag)
}

// ListMemories returns all memories, optionally filtered by project
func (s *Store) ListMemories(project string) ([]*Memory, error) {
	return s.db.ListMemories(project)
}

// Delete removes a memory by ID
func (s *Store) Delete(id string) error {
	return s.db.DeleteMemory(id)
}

// Graph queries the knowledge graph from a memory
func (s *Store) Graph(fromMemoryID string, depth int) ([]*Memory, error) {
	if depth <= 0 {
		depth = 2
	}

	visited := make(map[string]bool)
	var results []*Memory

	// BFS traversal of the graph
	queue := []string{fromMemoryID}
	currentDepth := 0

	for currentDepth < depth && len(queue) > 0 {
		var nextQueue []string
		for _, id := range queue {
			if visited[id] {
				continue
			}
			visited[id] = true

			// Get the memory itself
			m, _, err := s.db.GetMemory(id)
			if err != nil || m == nil {
				continue
			}
			results = append(results, m)

			// Get connected memories
			edges, err := s.db.GetEdges(id)
			if err != nil {
				continue
			}
			for _, edge := range edges {
				if !visited[edge.ToID] {
					nextQueue = append(nextQueue, edge.ToID)
				}
				if !visited[edge.FromID] {
					nextQueue = append(nextQueue, edge.FromID)
				}
			}
		}
		queue = nextQueue
		currentDepth++
	}

	return results, nil
}

// AddEdge adds an edge between two memories
func (s *Store) AddEdge(fromID, toID, relation string, weight float64) error {
	if weight == 0 {
		weight = 1.0
	}
	return s.db.AddEdge(fromID, toID, relation, weight)
}

// GetEdges returns all edges connected to a memory
func (s *Store) GetEdges(memoryID string) ([]*MemoryEdge, error) {
	return s.db.GetEdges(memoryID)
}

// Prune removes memories not matching keep tags
func (s *Store) Prune(keepTags []string) (int, error) {
	if len(keepTags) == 0 {
		keepTags = s.config.KeepTags
	}

	count, err := s.db.PruneMemories(keepTags)
	if err != nil {
		return 0, fmt.Errorf("failed to prune memories: %w", err)
	}

	logger.Info("Pruning completed", "deleted", count, "kept_tags", keepTags)
	return count, nil
}

// Encrypt encrypts the database with the given key file
func (s *Store) Encrypt(keyFile string) error {
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Close current database
	s.db.Close()

	// Encrypt the database file
	if err := s.enc.EncryptFile(s.config.DBPath, key); err != nil {
		return fmt.Errorf("failed to encrypt database: %w", err)
	}

	logger.Info("Database encrypted successfully")
	return nil
}

// Decrypt decrypts the database with the given key file
func (s *Store) Decrypt(keyFile string) error {
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Close current database
	s.db.Close()

	// Decrypt the database file
	plaintextPath := s.config.DBPath + ".plain"
	if err := s.enc.DecryptFile(s.config.DBPath, plaintextPath, key); err != nil {
		return fmt.Errorf("failed to decrypt database: %w", err)
	}

	// Replace encrypted file with decrypted
	if err := os.Rename(plaintextPath, s.config.DBPath); err != nil {
		return fmt.Errorf("failed to replace database: %w", err)
	}

	logger.Info("Database decrypted successfully")
	return nil
}

// SyncPush pushes local changes to remote
func (s *Store) SyncPush(remote string) error {
	if remote == "" {
		remote = s.config.Remote
	}
	return s.sync.Push(remote)
}

// SyncPull pulls changes from remote
func (s *Store) SyncPull(remote string) error {
	if remote == "" {
		remote = s.config.Remote
	}
	return s.sync.Pull(remote)
}

// VectorSearcher interface for vector search implementations
type VectorSearcher interface {
	GenerateEmbedding(text string) ([]float32, error)
	Search(query []float32, vectors [][]float32, k int) ([]int, []float64, error)
}

// SimpleVectorSearch provides a simple embedding-based search
type SimpleVectorSearch struct{}

// NewSimpleVectorSearch creates a new simple vector search
func NewSimpleVectorSearch() *SimpleVectorSearch {
	return &SimpleVectorSearch{}
}

// GenerateEmbedding generates a simple embedding using word frequency
// This is a placeholder that should be replaced with a proper embedding model
func (svs *SimpleVectorSearch) GenerateEmbedding(text string) ([]float32, error) {
	// Simple TF-IDF-like embedding for demonstration
	// In production, this should use a proper embedding model
	words := make(map[string]float32)
	wordCount := float32(0)

	// Simple tokenization
	var tokens []string
	start := 0
	for i, r := range text {
		if r == ' ' || r == '\n' || r == '\t' {
			if i > start {
				tokens = append(tokens, text[start:i])
			}
			start = i + 1
		}
	}
	if start < len(text) {
		tokens = append(tokens, text[start:])
	}

	for _, word := range tokens {
		words[word]++
		wordCount++
	}

	// Normalize
	dim := 256 // Fixed dimension for simplicity
	embedding := make([]float32, dim)

	i := 0
	for _, count := range words {
		if i >= dim {
			break
		}
		embedding[i] = count / wordCount
		i++
	}

	// Simple hash-based additional dimensions
	for idx, r := range text {
		if idx >= dim {
			break
		}
		embedding[idx] += float32(r) / 256.0
	}

	return embedding, nil
}

// Search performs cosine similarity search
func (svs *SimpleVectorSearch) Search(query []float32, vectors [][]float32, k int) ([]int, []float64, error) {
	if len(vectors) == 0 {
		return nil, nil, nil
	}

	type result struct {
		index    int
		distance float64
	}

	// Compute cosine similarity
	results := make([]result, 0, len(vectors))
	queryNorm := cosineNorm(query)

	for i, vec := range vectors {
		if len(vec) != len(query) {
			continue
		}
		similarity := cosineSimilarity(query, vec, queryNorm)
		results = append(results, result{i, similarity})
	}

	// Sort by similarity (descending)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].distance > results[i].distance {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Take top k
	if k > len(results) {
		k = len(results)
	}
	indices := make([]int, k)
	distances := make([]float64, k)
	for i := 0; i < k; i++ {
		indices[i] = results[i].index
		distances[i] = results[i].distance
	}

	return indices, distances, nil
}

func cosineNorm(vec []float32) float64 {
	sum := 0.0
	for _, v := range vec {
		sum += float64(v * v)
	}
	return math.Sqrt(sum)
}

func cosineSimilarity(a, b []float32, aNorm float64) float64 {
	dot := 0.0
	for i := range a {
		dot += float64(a[i] * b[i])
	}
	if aNorm == 0 {
		return 0
	}
	return dot / aNorm
}
