// Package memory provides the Mimir memory engine for ODIN
package memory

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection
type DB struct {
	db *sql.DB
}

// VectorSearchResult represents a vector search result from the database
type VectorSearchResult struct {
	MemoryID string
	Score    float64
	Distance float64
}

// NewDB creates a new database connection and initializes schema
func NewDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	d := &DB{db: db}
	if err := d.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return d, nil
}

// initSchema creates the necessary tables
func (d *DB) initSchema() error {
	schema := `
	-- Memories table
	CREATE TABLE IF NOT EXISTS memories (
		id          TEXT PRIMARY KEY,
		content     TEXT NOT NULL,
		project     TEXT DEFAULT '',
		tags        TEXT DEFAULT '[]',
		metadata    TEXT DEFAULT '{}',
		encrypted   INTEGER DEFAULT 0,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Knowledge graph edges
	CREATE TABLE IF NOT EXISTS memory_edges (
		from_id     TEXT NOT NULL,
		to_id       TEXT NOT NULL,
		relation    TEXT NOT NULL,
		weight      REAL DEFAULT 1.0,
		PRIMARY KEY (from_id, to_id, relation),
		FOREIGN KEY (from_id) REFERENCES memories(id) ON DELETE CASCADE,
		FOREIGN KEY (to_id) REFERENCES memories(id) ON DELETE CASCADE
	);

	-- Memory vectors for semantic search
	CREATE TABLE IF NOT EXISTS memory_vectors (
		memory_id   TEXT PRIMARY KEY,
		embedding   BLOB NOT NULL,
		FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE
	);

	-- Full-text search using FTS5
	CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
		content,
		content='memories',
		content_rowid='rowid'
	);

	-- Triggers to keep FTS5 in sync
	CREATE TRIGGER IF NOT EXISTS memories_fts_insert AFTER INSERT ON memories BEGIN
		INSERT INTO memories_fts(rowid, content) VALUES (NEW.rowid, NEW.content);
	END;

	CREATE TRIGGER IF NOT EXISTS memories_fts_delete AFTER DELETE ON memories BEGIN
		DELETE FROM memories_fts WHERE rowid = OLD.rowid;
	END;

	CREATE TRIGGER IF NOT EXISTS memories_fts_update AFTER UPDATE ON memories BEGIN
		DELETE FROM memories_fts WHERE rowid = OLD.rowid;
		INSERT INTO memories_fts(rowid, content) VALUES (NEW.rowid, NEW.content);
	END;

	-- Indexes for common queries
	CREATE INDEX IF NOT EXISTS idx_memories_project ON memories(project);
	CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at);
	CREATE INDEX IF NOT EXISTS idx_memories_tags ON memories(tags);
	CREATE INDEX IF NOT EXISTS idx_edges_from ON memory_edges(from_id);
	CREATE INDEX IF NOT EXISTS idx_edges_to ON memory_edges(to_id);
	`

	_, err := d.db.Exec(schema)
	return err
}

// Exec executes a SQL statement
func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

// Query executes a query that returns rows
func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// QueryRow executes a query that returns at most one row
func (d *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.db.Close()
}

// UpsertMemory inserts or updates a memory
func (d *DB) UpsertMemory(m *Memory, metadataJSON []byte) error {
	query := `
		INSERT INTO memories (id, content, project, tags, metadata, encrypted, created_at, updated_at, accessed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			content = excluded.content,
			project = excluded.project,
			tags = excluded.tags,
			metadata = excluded.metadata,
			encrypted = excluded.encrypted,
			updated_at = excluded.updated_at,
			accessed_at = excluded.accessed_at
	`

	_, err := d.db.Exec(query,
		m.ID,
		m.Content,
		m.Project,
		strings.Join(m.Tags, ","),
		string(metadataJSON),
		m.Encrypted,
		m.CreatedAt,
		m.UpdatedAt,
		m.AccessedAt,
	)
	return err
}

// GetMemory retrieves a memory by ID
func (d *DB) GetMemory(id string) (*Memory, []byte, error) {
	query := `
		SELECT id, content, project, tags, metadata, encrypted, created_at, updated_at, accessed_at
		FROM memories WHERE id = ?
	`

	var m Memory
	var tagsStr string
	var metadataStr string

	err := d.db.QueryRow(query, id).Scan(
		&m.ID,
		&m.Content,
		&m.Project,
		&tagsStr,
		&metadataStr,
		&m.Encrypted,
		&m.CreatedAt,
		&m.UpdatedAt,
		&m.AccessedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	// Parse tags
	if tagsStr != "" {
		m.Tags = strings.Split(tagsStr, ",")
	}

	return &m, []byte(metadataStr), nil
}

// DeleteMemory removes a memory by ID
func (d *DB) DeleteMemory(id string) error {
	_, err := d.db.Exec("DELETE FROM memories WHERE id = ?", id)
	return err
}

// UpdateAccessedAt updates the accessed_at timestamp
func (d *DB) UpdateAccessedAt(id string) error {
	_, err := d.db.Exec("UPDATE memories SET accessed_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

// StoreVector stores a vector embedding for a memory
func (d *DB) StoreVector(memoryID string, embedding []float32) error {
	// Convert float32 to bytes
	bytes := make([]byte, len(embedding)*4)
	for i, v := range embedding {
		binary.LittleEndian.PutUint32(bytes[i*4:], math.Float32bits(v))
	}

	query := `
		INSERT INTO memory_vectors (memory_id, embedding)
		VALUES (?, ?)
		ON CONFLICT(memory_id) DO UPDATE SET embedding = excluded.embedding
	`
	_, err := d.db.Exec(query, memoryID, bytes)
	return err
}

// GetVector retrieves a vector embedding for a memory
func (d *DB) GetVector(memoryID string) ([]float32, error) {
	var bytes []byte
	err := d.db.QueryRow("SELECT embedding FROM memory_vectors WHERE memory_id = ?", memoryID).Scan(&bytes)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Convert bytes to float32
	embedding := make([]float32, len(bytes)/4)
	for i := 0; i < len(embedding); i++ {
		embedding[i] = math.Float32frombits(binary.LittleEndian.Uint32(bytes[i*4:]))
	}

	return embedding, nil
}

// VectorSearch performs semantic search using stored vectors
func (d *DB) VectorSearch(queryEmbedding []float32, limit int) ([]VectorSearchResult, error) {
	rows, err := d.db.Query("SELECT memory_id, embedding FROM memory_vectors")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type candidate struct {
		memoryID string
		vector   []float32
	}

	var candidates []candidate
	for rows.Next() {
		var memoryID string
		var bytes []byte
		if err := rows.Scan(&memoryID, &bytes); err != nil {
			continue
		}

		vector := make([]float32, len(bytes)/4)
		for i := 0; i < len(vector); i++ {
			vector[i] = math.Float32frombits(binary.LittleEndian.Uint32(bytes[i*4:]))
		}
		candidates = append(candidates, candidate{memoryID, vector})
	}

	// Compute similarities
	type result struct {
		memoryID string
		score    float64
	}

	var results []result
	for _, c := range candidates {
		score := cosineSim(queryEmbedding, c.vector)
		results = append(results, result{c.memoryID, score})
	}

	// Sort by score descending
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Take top k
	if limit > len(results) {
		limit = len(results)
	}
	vectorResults := make([]VectorSearchResult, limit)
	for i := 0; i < limit; i++ {
		vectorResults[i] = VectorSearchResult{
			MemoryID: results[i].memoryID,
			Score:    results[i].score,
			Distance: 1.0 - results[i].score, // Convert similarity to distance
		}
	}

	return vectorResults, nil
}

// FTSSearch performs full-text search using FTS5
func (d *DB) FTSSearch(query string, limit int) ([]*Memory, error) {
	searchQuery := `
		SELECT m.id, m.content, m.project, m.tags, m.metadata, m.encrypted, 
			   m.created_at, m.updated_at, m.accessed_at
		FROM memories m
		INNER JOIN memories_fts fts ON m.rowid = fts.rowid
		WHERE memories_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`

	rows, err := d.db.Query(searchQuery, query, limit)
	if err != nil {
		// Fallback to LIKE search if FTS5 fails
		return d.likeSearch(query, limit)
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		var m Memory
		var tagsStr string
		var metadataStr string

		err := rows.Scan(
			&m.ID,
			&m.Content,
			&m.Project,
			&tagsStr,
			&metadataStr,
			&m.Encrypted,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.AccessedAt,
		)
		if err != nil {
			continue
		}

		if tagsStr != "" {
			m.Tags = strings.Split(tagsStr, ",")
		}

		memories = append(memories, &m)
	}

	return memories, nil
}

// likeSearch performs a simple LIKE search as fallback
func (d *DB) likeSearch(query string, limit int) ([]*Memory, error) {
	searchQuery := `
		SELECT id, content, project, tags, metadata, encrypted, 
			   created_at, updated_at, accessed_at
		FROM memories
		WHERE content LIKE ?
		ORDER BY accessed_at DESC
		LIMIT ?
	`

	rows, err := d.db.Query(searchQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		var m Memory
		var tagsStr string
		var metadataStr string

		err := rows.Scan(
			&m.ID,
			&m.Content,
			&m.Project,
			&tagsStr,
			&metadataStr,
			&m.Encrypted,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.AccessedAt,
		)
		if err != nil {
			continue
		}

		if tagsStr != "" {
			m.Tags = strings.Split(tagsStr, ",")
		}

		memories = append(memories, &m)
	}

	return memories, nil
}

// ListTags returns all unique tags
func (d *DB) ListTags() ([]string, error) {
	rows, err := d.db.Query("SELECT DISTINCT tags FROM memories WHERE tags != ''")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagSet := make(map[string]bool)
	for rows.Next() {
		var tagsStr string
		if err := rows.Scan(&tagsStr); err != nil {
			continue
		}
		for _, tag := range strings.Split(tagsStr, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagSet[tag] = true
			}
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags, nil
}

// AddTag adds a tag to a memory
func (d *DB) AddTag(memoryID, tag string) error {
	m, _, err := d.GetMemory(memoryID)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("memory not found: %s", memoryID)
	}

	// Check if tag already exists
	for _, t := range m.Tags {
		if t == tag {
			return nil
		}
	}

	m.Tags = append(m.Tags, tag)
	_, err = d.db.Exec("UPDATE memories SET tags = ? WHERE id = ?", strings.Join(m.Tags, ","), memoryID)
	return err
}

// ListMemories returns all memories, optionally filtered by project
func (d *DB) ListMemories(project string) ([]*Memory, error) {
	var query string
	var args []interface{}

	if project != "" {
		query = `
			SELECT id, content, project, tags, metadata, encrypted, 
				   created_at, updated_at, accessed_at
			FROM memories WHERE project = ?
			ORDER BY accessed_at DESC
		`
		args = []interface{}{project}
	} else {
		query = `
			SELECT id, content, project, tags, metadata, encrypted, 
				   created_at, updated_at, accessed_at
			FROM memories
			ORDER BY accessed_at DESC
		`
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		var m Memory
		var tagsStr string
		var metadataStr string

		err := rows.Scan(
			&m.ID,
			&m.Content,
			&m.Project,
			&tagsStr,
			&metadataStr,
			&m.Encrypted,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.AccessedAt,
		)
		if err != nil {
			continue
		}

		if tagsStr != "" {
			m.Tags = strings.Split(tagsStr, ",")
		}

		memories = append(memories, &m)
	}

	return memories, nil
}

// AddEdge adds an edge between two memories
func (d *DB) AddEdge(fromID, toID, relation string, weight float64) error {
	query := `
		INSERT INTO memory_edges (from_id, to_id, relation, weight)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(from_id, to_id, relation) DO UPDATE SET weight = excluded.weight
	`
	_, err := d.db.Exec(query, fromID, toID, relation, weight)
	return err
}

// GetEdges returns all edges connected to a memory
func (d *DB) GetEdges(memoryID string) ([]*MemoryEdge, error) {
	query := `
		SELECT from_id, to_id, relation, weight
		FROM memory_edges
		WHERE from_id = ? OR to_id = ?
	`

	rows, err := d.db.Query(query, memoryID, memoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []*MemoryEdge
	for rows.Next() {
		var e MemoryEdge
		if err := rows.Scan(&e.FromID, &e.ToID, &e.Relation, &e.Weight); err != nil {
			continue
		}
		edges = append(edges, &e)
	}

	return edges, nil
}

// PruneMemories removes memories not matching keep tags
func (d *DB) PruneMemories(keepTags []string) (int, error) {
	if len(keepTags) == 0 {
		return 0, nil
	}

	// Get all memories
	memories, err := d.ListMemories("")
	if err != nil {
		return 0, err
	}

	// Find memories to delete (those without any keep tag)
	toDelete := make([]string, 0)
	for _, m := range memories {
		hasProtectedTag := false
		for _, memTag := range m.Tags {
			for _, keepTag := range keepTags {
				if strings.EqualFold(strings.TrimSpace(memTag), keepTag) {
					hasProtectedTag = true
					break
				}
			}
			if hasProtectedTag {
				break
			}
		}
		if !hasProtectedTag {
			toDelete = append(toDelete, m.ID)
		}
	}

	// Delete memories
	count := 0
	for _, id := range toDelete {
		if err := d.DeleteMemory(id); err == nil {
			count++
		}
	}

	return count, nil
}

// GetAllMemories returns all memories for sync purposes
func (d *DB) GetAllMemories() ([]*Memory, error) {
	query := `
		SELECT id, content, project, tags, metadata, encrypted, 
			   created_at, updated_at, accessed_at
		FROM memories
		ORDER BY created_at
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		var m Memory
		var tagsStr string
		var metadataStr string

		err := rows.Scan(
			&m.ID,
			&m.Content,
			&m.Project,
			&tagsStr,
			&metadataStr,
			&m.Encrypted,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.AccessedAt,
		)
		if err != nil {
			continue
		}

		if tagsStr != "" {
			m.Tags = strings.Split(tagsStr, ",")
		}
		if metadataStr != "" && metadataStr != "{}" {
			json.Unmarshal([]byte(metadataStr), &m.Metadata)
		}

		memories = append(memories, &m)
	}

	return memories, nil
}

// GetAllEdges returns all edges for sync purposes
func (d *DB) GetAllEdges() ([]*MemoryEdge, error) {
	query := `SELECT from_id, to_id, relation, weight FROM memory_edges`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []*MemoryEdge
	for rows.Next() {
		var e MemoryEdge
		if err := rows.Scan(&e.FromID, &e.ToID, &e.Relation, &e.Weight); err != nil {
			continue
		}
		edges = append(edges, &e)
	}

	return edges, nil
}

// ImportMemories imports memories from another database (for sync)
func (d *DB) ImportMemories(memories []*Memory) error {
	for _, m := range memories {
		metadataJSON, _ := json.Marshal(m.Metadata)
		if err := d.UpsertMemory(m, metadataJSON); err != nil {
			return err
		}
	}
	return nil
}

// ImportEdges imports edges from another database (for sync)
func (d *DB) ImportEdges(edges []*MemoryEdge) error {
	for _, e := range edges {
		if err := d.AddEdge(e.FromID, e.ToID, e.Relation, e.Weight); err != nil {
			return err
		}
	}
	return nil
}

// cosineSim computes cosine similarity between two vectors
func cosineSim(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	dot := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		dot += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
