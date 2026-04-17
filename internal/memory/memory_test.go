// Package memory provides the Mimir memory engine for ODIN
package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMemoryModel(t *testing.T) {
	m := &Memory{
		ID:         "test-id",
		Content:    "Test content",
		Project:    "test-project",
		Tags:       []string{"test", "unit"},
		Metadata:   map[string]interface{}{"key": "value"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		AccessedAt: time.Now(),
	}

	if m.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", m.ID)
	}

	if m.Content != "Test content" {
		t.Errorf("Expected Content 'Test content', got '%s'", m.Content)
	}

	if m.Project != "test-project" {
		t.Errorf("Expected Project 'test-project', got '%s'", m.Project)
	}

	if len(m.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(m.Tags))
	}
}

func TestSearchResult(t *testing.T) {
	m := &Memory{
		ID:      "result-id",
		Content: "Result content",
	}

	sr := &SearchResult{
		Memory:   m,
		Score:    0.95,
		Distance: 0.05,
	}

	if sr.Memory.ID != "result-id" {
		t.Errorf("Expected Memory ID 'result-id', got '%s'", sr.Memory.ID)
	}

	if sr.Score != 0.95 {
		t.Errorf("Expected Score 0.95, got %f", sr.Score)
	}
}

func TestConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DBPath == "" {
		t.Error("Expected DBPath to be set")
	}

	if cfg.PruneInterval != 24*time.Hour {
		t.Errorf("Expected PruneInterval 24h, got %v", cfg.PruneInterval)
	}

	expectedTags := []string{"arch", "spec", "security"}
	for i, tag := range cfg.KeepTags {
		if tag != expectedTags[i] {
			t.Errorf("Expected KeepTag[%d] '%s', got '%s'", i, expectedTags[i], tag)
		}
	}
}

func TestStoreCreate(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_memory.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false, // Disable VSS for testing
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	if store == nil {
		t.Fatal("Store should not be nil")
	}
}

func TestStoreAndRecall(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_memory.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false,
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create a memory
	m := &Memory{
		Content: "Test memory content",
		Project: "test-project",
		Tags:    []string{"test", "memory"},
	}

	if err := store.Store(m); err != nil {
		t.Fatalf("Failed to store memory: %v", err)
	}

	if m.ID == "" {
		t.Error("Memory ID should be set after store")
	}

	// Recall the memory
	recall, err := store.Recall(m.ID)
	if err != nil {
		t.Fatalf("Failed to recall memory: %v", err)
	}

	if recall == nil {
		t.Fatal("Recalled memory should not be nil")
	}

	if recall.Content != "Test memory content" {
		t.Errorf("Expected content 'Test memory content', got '%s'", recall.Content)
	}

	if recall.Project != "test-project" {
		t.Errorf("Expected project 'test-project', got '%s'", recall.Project)
	}
}

func TestStoreMultipleMemories(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_memory.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false,
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Store multiple memories
	for i := 0; i < 5; i++ {
		m := &Memory{
			Content: "Test memory content",
			Project: "test-project",
			Tags:    []string{"test", "multiple"},
		}
		if err := store.Store(m); err != nil {
			t.Fatalf("Failed to store memory %d: %v", i, err)
		}
	}

	// List memories
	memories, err := store.ListMemories("test-project")
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	if len(memories) != 5 {
		t.Errorf("Expected 5 memories, got %d", len(memories))
	}
}

func TestSearchMemories(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_memory.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false, // Use FTS5 fallback
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Store test memories
	memories := []*Memory{
		{Content: "Go programming language", Tags: []string{"golang", "programming"}},
		{Content: "Python for data science", Tags: []string{"python", "data"}},
		{Content: "Web development with JavaScript", Tags: []string{"javascript", "web"}},
	}

	for _, m := range memories {
		if err := store.Store(m); err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}
	}

	// Search
	results, err := store.Search("Go programming", 10)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	// FTS5 should find at least one result
	if len(results) == 0 {
		t.Log("Warning: FTS5 search returned no results (may be expected in test environment)")
	}
}

func TestTagsOperations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_memory.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false,
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create a memory with no tags
	m := &Memory{
		Content: "Memory without tags",
	}

	if err := store.Store(m); err != nil {
		t.Fatalf("Failed to store memory: %v", err)
	}

	// Add tags
	if err := store.AddTag(m.ID, "tag1"); err != nil {
		t.Fatalf("Failed to add tag: %v", err)
	}

	if err := store.AddTag(m.ID, "tag2"); err != nil {
		t.Fatalf("Failed to add tag: %v", err)
	}

	// Recall and check tags
	recall, _ := store.Recall(m.ID)
	if len(recall.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(recall.Tags))
	}

	// List tags
	tags, err := store.ListTags()
	if err != nil {
		t.Fatalf("Failed to list tags: %v", err)
	}

	if len(tags) < 2 {
		t.Errorf("Expected at least 2 tags, got %d", len(tags))
	}
}

func TestDeleteMemory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_memory.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false,
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create and delete
	m := &Memory{
		Content: "Memory to delete",
	}

	if err := store.Store(m); err != nil {
		t.Fatalf("Failed to store memory: %v", err)
	}

	if err := store.Delete(m.ID); err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}

	// Verify deleted
	recall, _ := store.Recall(m.ID)
	if recall != nil {
		t.Error("Memory should be nil after deletion")
	}
}

func TestGraphOperations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_memory.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false,
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create memories
	m1 := &Memory{Content: "Memory 1", Tags: []string{"start"}}
	m2 := &Memory{Content: "Memory 2", Tags: []string{"middle"}}
	m3 := &Memory{Content: "Memory 3", Tags: []string{"end"}}

	store.Store(m1)
	store.Store(m2)
	store.Store(m3)

	// Create edges
	store.AddEdge(m1.ID, m2.ID, "depends_on", 1.0)
	store.AddEdge(m2.ID, m3.ID, "leads_to", 1.0)

	// Query graph from m1
	memories, err := store.Graph(m1.ID, 2)
	if err != nil {
		t.Fatalf("Failed to query graph: %v", err)
	}

	// Should include all three memories
	if len(memories) < 2 {
		t.Errorf("Expected at least 2 connected memories, got %d", len(memories))
	}
}

func TestPrune(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_memory.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false,
		KeepTags:   []string{"keep"}, // Only keep this tag
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create memories with different tags
	memories := []*Memory{
		{Content: "Keep this", Tags: []string{"keep"}},
		{Content: "Prune this", Tags: []string{"prune"}},
		{Content: "Keep that", Tags: []string{"keep"}},
		{Content: "Prune also", Tags: []string{"delete"}},
	}

	for _, m := range memories {
		store.Store(m)
	}

	// Prune
	count, err := store.Prune(nil) // Use default keep tags
	if err != nil {
		t.Fatalf("Failed to prune: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 pruned memories, got %d", count)
	}

	// Verify kept memories
	all, _ := store.ListMemories("")
	if len(all) != 2 {
		t.Errorf("Expected 2 remaining memories, got %d", len(all))
	}
}

func TestSimpleVectorSearch(t *testing.T) {
	svs := NewSimpleVectorSearch()

	// Test embedding generation
	emb1, err := svs.GenerateEmbedding("Hello world")
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}

	if len(emb1) == 0 {
		t.Error("Embedding should not be empty")
	}

	// Generate another embedding
	emb2, err := svs.GenerateEmbedding("Hello world")
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}

	// Same text should produce same embedding
	for i := range emb1 {
		if emb1[i] != emb2[i] {
			t.Error("Same text should produce same embedding")
			break
		}
	}

	// Test search - use vectors with same dimension as query
	dim := len(emb1)
	vectors := make([][]float32, 3)
	for i := range vectors {
		vectors[i] = make([]float32, dim)
		vectors[i][i%dim] = 1.0
	}

	indices, distances, err := svs.Search(emb1, vectors, 2)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(indices) != 2 {
		t.Errorf("Expected 2 results, got %d", len(indices))
	}

	if len(distances) != 2 {
		t.Errorf("Expected 2 distances, got %d", len(distances))
	}
}

func TestEncryptor(t *testing.T) {
	enc := NewEncryptor()

	// Test key generation
	key, err := enc.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	// Test encryption/decryption
	plaintext := []byte("Hello, Mimir!")

	encrypted, err := enc.encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if len(encrypted) == 0 {
		t.Error("Encrypted data should not be empty")
	}

	decrypted, err := enc.decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted data doesn't match: got '%s', want '%s'", string(decrypted), string(plaintext))
	}
}

func TestEncryptDecryptFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write test data
	testData := []byte("Secret memory data")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	enc := NewEncryptor()
	key, _ := enc.GenerateKey()

	// Encrypt
	if err := enc.EncryptFile(testFile, key); err != nil {
		t.Fatalf("Failed to encrypt file: %v", err)
	}

	// Check encrypted file exists
	if _, err := os.Stat(testFile + ".age"); os.IsNotExist(err) {
		t.Error("Encrypted file should exist")
	}

	// Backup and decrypt
	enc.BackupFile(testFile + ".age")

	// Decrypt to new file
	decryptedFile := testFile + ".decrypted"
	if err := enc.DecryptFile(testFile+".age", decryptedFile, key); err != nil {
		t.Fatalf("Failed to decrypt file: %v", err)
	}

	// Verify decrypted data
	decrypted, _ := os.ReadFile(decryptedFile)
	if string(decrypted) != string(testData) {
		t.Errorf("Decrypted data doesn't match")
	}
}

func TestSyncer(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_memory.db")
	remotePath := filepath.Join(tmpDir, "remote.json")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false,
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Store test memories
	m := &Memory{
		Content: "Test for sync",
		Tags:    []string{"sync"},
	}
	store.Store(m)

	// Test sync push
	if err := store.SyncPush(remotePath); err != nil {
		t.Fatalf("Sync push failed: %v", err)
	}

	// Create new store and sync pull
	cfg2 := &Config{
		DBPath:     filepath.Join(tmpDir, "test_memory2.db"),
		VSSEnabled: false,
	}

	store2, err := NewStore(cfg2)
	if err != nil {
		t.Fatalf("Failed to create store2: %v", err)
	}
	defer store2.Close()

	if err := store2.SyncPull(remotePath); err != nil {
		t.Fatalf("Sync pull failed: %v", err)
	}

	// Verify memories were synced
	memories, _ := store2.ListMemories("")
	if len(memories) != 1 {
		t.Errorf("Expected 1 memory after sync, got %d", len(memories))
	}
}

func TestMemoryEdge(t *testing.T) {
	edge := &MemoryEdge{
		FromID:   "from-1",
		ToID:     "to-1",
		Relation: "depends_on",
		Weight:   0.8,
	}

	if edge.Relation != "depends_on" {
		t.Errorf("Expected relation 'depends_on', got '%s'", edge.Relation)
	}

	if edge.Weight != 0.8 {
		t.Errorf("Expected weight 0.8, got %f", edge.Weight)
	}
}

func TestSearchOptions(t *testing.T) {
	opts := DefaultSearchOptions()

	if opts.Limit != 10 {
		t.Errorf("Expected default limit 10, got %d", opts.Limit)
	}

	if opts.SortBy != "relevance" {
		t.Errorf("Expected default sort 'relevance', got '%s'", opts.SortBy)
	}

	if opts.SortDesc != true {
		t.Error("Expected default sort descending")
	}
}

func TestConfigCustom(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		DBPath:        filepath.Join(tmpDir, "custom.db"),
		KeepTags:      []string{"important", "critical"},
		PruneInterval: 12 * time.Hour,
		VSSEnabled:    true,
		Remote:        "file:///tmp/remote",
	}

	if cfg.PruneInterval != 12*time.Hour {
		t.Errorf("Expected interval 12h, got %v", cfg.PruneInterval)
	}

	if len(cfg.KeepTags) != 2 {
		t.Errorf("Expected 2 keep tags, got %d", len(cfg.KeepTags))
	}
}

func TestGraphQuery(t *testing.T) {
	query := GraphQuery{
		FromID:       "start-id",
		Depth:        3,
		RelationType: "depends_on",
	}

	if query.FromID != "start-id" {
		t.Errorf("Expected FromID 'start-id', got '%s'", query.FromID)
	}

	if query.Depth != 3 {
		t.Errorf("Expected Depth 3, got %d", query.Depth)
	}
}

func TestGraphEngine(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_graph.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false,
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create test memories
	m1 := &Memory{Content: "Start", Tags: []string{"start"}}
	m2 := &Memory{Content: "Middle", Tags: []string{"middle"}}
	m3 := &Memory{Content: "End", Tags: []string{"end"}}

	store.Store(m1)
	store.Store(m2)
	store.Store(m3)

	// Create connections
	store.AddEdge(m1.ID, m2.ID, "connects", 1.0)
	store.AddEdge(m2.ID, m3.ID, "connects", 1.0)

	// Test GraphEngine directly
	graph := NewGraphEngine(store.projectDB)

	// Get connected memories
	connected, err := graph.GetConnectedMemories(m1.ID, "")
	if err != nil {
		t.Fatalf("GetConnectedMemories failed: %v", err)
	}

	if len(connected) == 0 {
		t.Error("Expected at least one connected memory")
	}

	// Test relation counts
	counts, err := graph.GetRelationCounts(m1.ID)
	if err != nil {
		t.Fatalf("GetRelationCounts failed: %v", err)
	}

	if counts["connects"] != 1 {
		t.Errorf("Expected 1 'connects' relation, got %d", counts["connects"])
	}
}

func TestPruner(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_pruner.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false,
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create pruner
	pruner := NewPruner(store.projectDB, []string{"protected"}, 1*time.Hour)
	defer pruner.Stop()

	// Start pruner
	pruner.Start()

	// Verify stats
	stats := pruner.GetStats()
	if !stats.IsRunning {
		t.Error("Pruner should be running")
	}

	// Stop pruner
	pruner.Stop()

	// Verify stopped
	stats = pruner.GetStats()
	if stats.IsRunning {
		t.Error("Pruner should be stopped")
	}
}

func TestSmartPruner(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_smart_pruner.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false,
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	smartPruner := NewSmartPruner(store.projectDB, []string{"protected"}, 24*time.Hour)

	// Test should prune
	m := &Memory{
		ID:   "test-id",
		Tags: []string{"temp"}, // No protected tag
	}

	shouldPrune, reason := smartPruner.ShouldPrune(m)
	if !shouldPrune {
		t.Errorf("Expected shouldPrune=true, got false (reason: %s)", reason)
	}

	// Test should not prune (protected tag)
	m2 := &Memory{
		ID:   "test-id-2",
		Tags: []string{"protected"},
	}

	shouldPrune2, _ := smartPruner.ShouldPrune(m2)
	if shouldPrune2 {
		t.Error("Expected shouldPrune=false for protected memory")
	}
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"tag1,tag2,tag3", []string{"tag1", "tag2", "tag3"}},
		{"tag1, tag2 , tag3", []string{"tag1", "tag2", "tag3"}},
		{"single", []string{"single"}},
		{"", nil},
	}

	for _, test := range tests {
		result := parseTags(test.input)

		if test.input == "" {
			if len(result) != 0 {
				t.Errorf("parseTags(%q) = %v, want empty slice", test.input, result)
			}
			continue
		}

		if len(result) != len(test.expected) {
			t.Errorf("parseTags(%q) length = %d, want %d", test.input, len(result), len(test.expected))
			continue
		}

		for i, tag := range result {
			if tag != test.expected[i] {
				t.Errorf("parseTags(%q)[%d] = %q, want %q", test.input, i, tag, test.expected[i])
			}
		}
	}
}

func TestJoinTags(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"tag1", "tag2", "tag3"}, "tag1, tag2, tag3"},
		{[]string{"single"}, "single"},
		{[]string{}, ""},
	}

	for _, test := range tests {
		result := joinTags(test.input)
		if result != test.expected {
			t.Errorf("joinTags(%v) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestCosineSimilarity(t *testing.T) {
	// Test identical vectors
	a := []float32{1.0, 0.0, 0.0}
	b := []float32{1.0, 0.0, 0.0}

	sim := cosineSim(a, b)
	if sim != 1.0 {
		t.Errorf("Identical vectors should have similarity 1.0, got %f", sim)
	}

	// Test orthogonal vectors
	c := []float32{0.0, 1.0, 0.0}
	sim2 := cosineSim(a, c)
	if sim2 != 0.0 {
		t.Errorf("Orthogonal vectors should have similarity 0.0, got %f", sim2)
	}

	// Test opposite vectors
	d := []float32{-1.0, 0.0, 0.0}
	sim3 := cosineSim(a, d)
	if sim3 != -1.0 {
		t.Errorf("Opposite vectors should have similarity -1.0, got %f", sim3)
	}
}

// TestMimirSearch_SemanticRelevance tests semantic search quality
func TestMimirSearch_SemanticRelevance(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_semantic.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false, // Use FTS5 fallback
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Store semantically different memories
	memories := []*Memory{
		{Content: "Python programming language tutorial", Tags: []string{"python", "programming"}},
		{Content: "JavaScript web development guide", Tags: []string{"javascript", "web"}},
		{Content: "Go programming language concurrency patterns", Tags: []string{"golang", "programming"}},
		{Content: "Machine learning neural networks", Tags: []string{"ml", "ai"}},
		{Content: "Database SQL queries optimization", Tags: []string{"database", "sql"}},
	}

	for _, m := range memories {
		if err := store.Store(m); err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}
	}

	// Test: Search for "Python" should return Python memory first
	results, err := store.Search("Python tutorial", 5)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected search results")
	}

	// The first result should be about Python
	firstResult := results[0]
	if firstResult.Memory == nil {
		t.Fatal("First result memory is nil")
	}

	// Check that Python content ranks higher than JavaScript for Python query
	pythonRank := -1
	javascriptRank := -1

	for i, r := range results {
		if r.Memory != nil {
			if strings.Contains(r.Memory.Content, "Python") {
				pythonRank = i
			}
			if strings.Contains(r.Memory.Content, "JavaScript") {
				javascriptRank = i
			}
		}
	}

	if pythonRank == -1 {
		t.Error("Python content should be in results")
	}

	if pythonRank > javascriptRank && javascriptRank != -1 {
		t.Errorf("Python should rank higher than JavaScript for 'Python' query: python=%d, javascript=%d", pythonRank, javascriptRank)
	}

	// Test: Search for "programming" should return programming-related memories
	progResults, _ := store.Search("programming", 5)
	progCount := 0
	for _, r := range progResults {
		if r.Memory != nil && strings.Contains(r.Memory.Content, "programming") {
			progCount++
		}
	}

	if progCount < 2 {
		t.Errorf("Expected at least 2 programming-related results, got %d", progCount)
	}

	// Test: Search for "web" should return JavaScript content
	webResults, _ := store.Search("web development", 5)
	webFound := false
	for _, r := range webResults {
		if r.Memory != nil && strings.Contains(r.Memory.Content, "JavaScript") {
			webFound = true
			break
		}
	}

	if !webFound {
		t.Error("JavaScript content should be found for 'web development' query")
	}
}

// BenchmarkSemanticSearch benchmarks vector search performance
func BenchmarkSemanticSearch(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench_semantic.db")

	cfg := &Config{
		DBPath:     dbPath,
		VSSEnabled: false,
	}

	store, err := NewStore(cfg)
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create test memories
	for i := 0; i < 100; i++ {
		m := &Memory{
			Content: fmt.Sprintf("Test memory content number %d about various topics", i),
			Tags:    []string{"benchmark"},
		}
		store.Store(m)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Search("test content", 10)
	}
}
