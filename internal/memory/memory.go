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
	DBPath        string // Project-specific DB
	GlobalDBPath  string // Global (cross-project) DB
	EncryptionKey string
	KeepTags      []string
	PruneInterval time.Duration
	Remote        string
	VSSEnabled    bool
	Embedder      Embedder // Custom embedder, nil means use DefaultEmbedder()
}

// DefaultConfig returns the default Mimir configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		DBPath:        filepath.Join(homeDir, ".odin", "memory.db"),
		GlobalDBPath:  filepath.Join(homeDir, ".odin", "global_memory.db"),
		EncryptionKey: "",
		KeepTags:      []string{"arch", "spec", "security"},
		PruneInterval: 24 * time.Hour,
		Remote:        "",
		VSSEnabled:    true,
	}
}

// Store represents the Mimir memory store
type Store struct {
	projectDB *DB
	globalDB  *DB
	config    *Config
	embedder  Embedder
	vss       VectorSearcher
	enc       *Encryptor
	sync      *Syncer
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

	projectDB, err := NewDB(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open project database: %w", err)
	}

	// Initialize global database if path is provided
	var globalDB *DB
	if cfg.GlobalDBPath != "" {
		gDir := filepath.Dir(cfg.GlobalDBPath)
		_ = os.MkdirAll(gDir, 0755)

		var gErr error
		globalDB, gErr = NewDB(cfg.GlobalDBPath)
		if gErr != nil {
			// Log warning but continue; global memory might not be available yet
			logger.Warn("Failed to open global database, working in project-only mode", "error", gErr)
		}
	}

	// Use custom embedder if provided, otherwise use DefaultEmbedder
	var embedder Embedder
	if cfg.Embedder != nil {
		embedder = cfg.Embedder
	} else {
		embedder = DefaultEmbedder()
	}

	store := &Store{
		projectDB: projectDB,
		globalDB:  globalDB,
		config:    cfg,
		embedder:  embedder,
		vss:       NewSimpleVectorSearch(),
		enc:       NewEncryptor(),
		sync:      NewSyncer(projectDB),
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

// initVectorSearch initializes the vector search engine in both databases
func (s *Store) initVectorSearch() error {
	// Create vector search table in project DB
	_, err := s.projectDB.Exec(`
		CREATE TABLE IF NOT EXISTS memory_vectors (
			memory_id TEXT PRIMARY KEY,
			embedding BLOB NOT NULL,
			FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create vectors table in project db: %w", err)
	}

	// Create vector search table in global DB if available
	if s.globalDB != nil {
		_, err = s.globalDB.Exec(`
			CREATE TABLE IF NOT EXISTS memory_vectors (
				memory_id TEXT PRIMARY KEY,
				embedding BLOB NOT NULL,
				FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE
			)
		`)
		if err != nil {
			logger.Warn("Failed to create vectors table in global db", "error", err)
		}
	}
	return nil
}

// Close closes all active database connections
func (s *Store) Close() error {
	var errs []error
	if s.projectDB != nil {
		if err := s.projectDB.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.globalDB != nil {
		if err := s.globalDB.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing databases: %v", errs)
	}
	return nil
}

// forEachDB runs a function on each active database
func (s *Store) forEachDB(fn func(*DB) error) error {
	var errs []error
	if s.projectDB != nil {
		if err := fn(s.projectDB); err != nil {
			errs = append(errs, err)
		}
	}
	if s.globalDB != nil {
		if err := fn(s.globalDB); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during database operation: %v", errs)
	}
	return nil
}

// Store saves a memory to the project store
func (s *Store) Store(m *Memory) error {
	return s.storeInternal(s.projectDB, m)
}

// StoreGlobal saves a memory to the global store
func (s *Store) StoreGlobal(m *Memory) error {
	if s.globalDB == nil {
		return fmt.Errorf("global memory not initialized")
	}
	return s.storeInternal(s.globalDB, m)
}

func (s *Store) storeInternal(db *DB, m *Memory) error {
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
	if s.vss != nil && len(m.Embedding) == 0 && s.embedder != nil {
		embedding, err := s.embedder.GenerateEmbedding(m.Content)
		if err != nil {
			logger.Warn("Failed to generate embedding", "error", err)
		} else {
			m.Embedding = embedding
			// Store vector
			if err := db.StoreVector(m.ID, embedding); err != nil {
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
	err := db.UpsertMemory(m, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to store memory: %w", err)
	}

	logger.Debug("Memory stored", "id", m.ID, "tags", m.Tags)
	return nil
}

// Recall retrieves a memory by ID from either store
func (s *Store) Recall(id string) (*Memory, error) {
	// Try project DB first
	m, metadataJSON, err := s.projectDB.GetMemory(id)
	if err == nil && m != nil {
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &m.Metadata)
		}
		s.projectDB.UpdateAccessedAt(id)
		return m, nil
	}

	// Try global DB
	if s.globalDB != nil {
		m, metadataJSON, err = s.globalDB.GetMemory(id)
		if err == nil && m != nil {
			if metadataJSON != nil {
				json.Unmarshal(metadataJSON, &m.Metadata)
			}
			s.globalDB.UpdateAccessedAt(id)
			return m, nil
		}
	}

	return nil, fmt.Errorf("failed to recall memory: not found in project or global store")
}

// Search performs semantic search on both project and global stores
func (s *Store) Search(query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	var allResults []SearchResult

	// Search project DB
	projectResults, err := s.searchInternal(s.projectDB, query, limit)
	if err == nil {
		allResults = append(allResults, projectResults...)
	}

	// Search global DB if available
	if s.globalDB != nil {
		globalResults, err := s.searchInternal(s.globalDB, query, limit)
		if err == nil {
			allResults = append(allResults, globalResults...)
		}
	}

	// Sort by score/distance and limit
	// (For now, just return combined list)
	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

func (s *Store) searchInternal(db *DB, query string, limit int) ([]SearchResult, error) {
	// If vector search is available, use it
	if s.vss != nil && s.embedder != nil {
		embedding, err := s.embedder.GenerateEmbedding(query)
		if err != nil {
			return nil, fmt.Errorf("failed to generate query embedding: %w", err)
		}

		results, err := db.VectorSearch(embedding, limit)
		if err != nil {
			return s.fts5SearchInternal(db, query, limit)
		}

		// Enrich results with full memory data
		searchResults := make([]SearchResult, 0, len(results))
		for _, r := range results {
			m, _, err := db.GetMemory(r.MemoryID)
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
	return s.fts5SearchInternal(db, query, limit)
}

// fts5SearchInternal performs FTS5 full-text search as fallback on a specific DB
func (s *Store) fts5SearchInternal(db *DB, query string, limit int) ([]SearchResult, error) {
	memories, err := db.FTSSearch(query, limit)
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

// ListTags returns all unique tags in the project store
func (s *Store) ListTags() ([]string, error) {
	return s.projectDB.ListTags()
}

// AddTag adds a tag to a memory in the project store
func (s *Store) AddTag(memoryID, tag string) error {
	return s.projectDB.AddTag(memoryID, tag)
}

// ListMemories returns all memories from the project store, optionally filtered by project
func (s *Store) ListMemories(project string) ([]*Memory, error) {
	return s.projectDB.ListMemories(project)
}

// Delete removes a memory by ID from the project store
func (s *Store) Delete(id string) error {
	return s.projectDB.DeleteMemory(id)
}

// Graph queries the knowledge graph from a memory
func (s *Store) Graph(fromMemoryID string, depth int) ([]*Memory, error) {
	if depth <= 0 {
		depth = 2
	}

	visited := make(map[string]bool)
	var results []*Memory

	// BFS traversal of the graph across both layers
	queue := []string{fromMemoryID}
	currentDepth := 0

	for currentDepth < depth && len(queue) > 0 {
		var nextQueue []string
		for _, id := range queue {
			if visited[id] {
				continue
			}
			visited[id] = true

			// Get the memory from either layer
			m, err := s.Recall(id)
			if err != nil || m == nil {
				continue
			}
			results = append(results, m)

			// Get connected memories from both layers
			// Edges in project DB
			if s.projectDB != nil {
				edges, _ := s.projectDB.GetEdges(id)
				for _, edge := range edges {
					if !visited[edge.ToID] {
						nextQueue = append(nextQueue, edge.ToID)
					}
					if !visited[edge.FromID] {
						nextQueue = append(nextQueue, edge.FromID)
					}
				}
			}
			// Edges in global DB
			if s.globalDB != nil {
				edges, _ := s.globalDB.GetEdges(id)
				for _, edge := range edges {
					if !visited[edge.ToID] {
						nextQueue = append(nextQueue, edge.ToID)
					}
					if !visited[edge.FromID] {
						nextQueue = append(nextQueue, edge.FromID)
					}
				}
			}
		}
		queue = nextQueue
		currentDepth++
	}

	return results, nil
}

// AddEdge adds an edge between two memories in the project store
func (s *Store) AddEdge(fromID, toID, relation string, weight float64) error {
	if weight == 0 {
		weight = 1.0
	}
	return s.projectDB.AddEdge(fromID, toID, relation, weight)
}

// GetEdges returns all edges connected to a memory in the project store
func (s *Store) GetEdges(memoryID string) ([]*MemoryEdge, error) {
	return s.projectDB.GetEdges(memoryID)
}

// Prune removes memories that don't match keep tags from project database
// (By default we don't prune global memory unless explicitly requested)
func (s *Store) Prune(keepTags []string) (int, error) {
	if s.projectDB == nil {
		return 0, nil
	}

	// Falls back to config if nil
	if len(keepTags) == 0 {
		keepTags = s.config.KeepTags
	}

	// Use simple pruning for now as expected by standard tests
	return s.projectDB.PruneMemories(keepTags)
}

// Encrypt encrypts the database with the given key file
func (s *Store) Encrypt(keyFile string) error {
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Close current databases
	s.Close()

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

	// Close current databases
	s.Close()

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
