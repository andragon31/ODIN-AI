// Package memory provides the Mimir memory engine for ODIN
package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/odin-ai/odin/pkg/logger"
)

// Syncer handles CRDT-based synchronization
type Syncer struct {
	db *DB
}

// NewSyncer creates a new Syncer
func NewSyncer(db *DB) *Syncer {
	return &Syncer{db: db}
}

// SyncState represents the synchronization state
type SyncState struct {
	LastSync    time.Time `json:"last_sync"`
	LocalCursor string    `json:"local_cursor"`
	RemoteURL   string    `json:"remote_url"`
}

// RemoteMemories represents memories fetched from remote
type RemoteMemories struct {
	Memories []*Memory     `json:"memories"`
	Edges    []*MemoryEdge `json:"edges"`
	State    SyncState     `json:"state"`
}

// Push pushes local changes to remote
func (s *Syncer) Push(remote string) error {
	if remote == "" {
		return fmt.Errorf("remote URL is required")
	}

	// Get all local memories and edges
	memories, err := s.db.GetAllMemories()
	if err != nil {
		return fmt.Errorf("failed to get local memories: %w", err)
	}

	edges, err := s.db.GetAllEdges()
	if err != nil {
		return fmt.Errorf("failed to get local edges: %w", err)
	}

	// Create remote package
	remoteData := RemoteMemories{
		Memories: memories,
		Edges:    edges,
		State: SyncState{
			LastSync:    time.Now(),
			LocalCursor: generateCursor(memories),
			RemoteURL:   remote,
		},
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(remoteData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize sync data: %w", err)
	}

	// Write to sync file
	syncPath := s.getSyncFilePath(remote)
	if err := os.WriteFile(syncPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write sync file: %w", err)
	}

	// In a real implementation, this would push to a git remote or cloud storage
	// For now, we'll implement a file-based sync mechanism
	if err := s.pushToRemote(remote, data); err != nil {
		return fmt.Errorf("failed to push to remote: %w", err)
	}

	logger.Info("Push completed", "memories", len(memories), "edges", len(edges), "remote", remote)
	return nil
}

// pushToRemote pushes data to the remote storage
func (s *Syncer) pushToRemote(remote string, data []byte) error {
	// Create remote directory if it doesn't exist
	remoteDir := filepath.Dir(remote)
	if err := os.MkdirAll(remoteDir, 0755); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Write to remote
	if err := os.WriteFile(remote, data, 0644); err != nil {
		return fmt.Errorf("failed to write to remote: %w", err)
	}

	return nil
}

// Pull pulls changes from remote
func (s *Syncer) Pull(remote string) error {
	if remote == "" {
		return fmt.Errorf("remote URL is required")
	}

	// Read from remote
	data, err := os.ReadFile(remote)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("Remote has no data yet, skipping pull")
			return nil
		}
		return fmt.Errorf("failed to read from remote: %w", err)
	}

	// Deserialize
	var remoteData RemoteMemories
	if err := json.Unmarshal(data, &remoteData); err != nil {
		return fmt.Errorf("failed to deserialize sync data: %w", err)
	}

	// Merge with local using CRDT-style merge
	if err := s.merge(remoteData); err != nil {
		return fmt.Errorf("failed to merge remote data: %w", err)
	}

	logger.Info("Pull completed",
		"memories", len(remoteData.Memories),
		"edges", len(remoteData.Edges),
		"remote", remote)
	return nil
}

// merge merges remote data with local data using CRDT semantics
func (s *Syncer) merge(remote RemoteMemories) error {
	// Get local data
	localMemories, err := s.db.GetAllMemories()
	if err != nil {
		return fmt.Errorf("failed to get local memories: %w", err)
	}

	localEdges, err := s.db.GetAllEdges()
	if err != nil {
		return fmt.Errorf("failed to get local edges: %w", err)
	}

	// Create maps for efficient lookup
	localMemoryMap := make(map[string]*Memory)
	for _, m := range localMemories {
		localMemoryMap[m.ID] = m
	}

	localEdgeMap := make(map[string]*MemoryEdge)
	for _, e := range localEdges {
		key := edgeKey(e.FromID, e.ToID, e.Relation)
		localEdgeMap[key] = e
	}

	// Merge memories - last-write-wins CRDT
	for _, remoteMem := range remote.Memories {
		localMem, exists := localMemoryMap[remoteMem.ID]
		if !exists {
			// New memory from remote
			if err := s.db.ImportMemories([]*Memory{remoteMem}); err != nil {
				logger.Warn("Failed to import memory", "id", remoteMem.ID, "error", err)
			}
		} else {
			// Compare timestamps - newer wins (LWW - Last Write Wins)
			if remoteMem.UpdatedAt.After(localMem.UpdatedAt) {
				// Remote is newer, update local
				if err := s.db.ImportMemories([]*Memory{remoteMem}); err != nil {
					logger.Warn("Failed to update memory", "id", remoteMem.ID, "error", err)
				}
			}
		}
	}

	// Merge edges - add missing edges
	for _, remoteEdge := range remote.Edges {
		key := edgeKey(remoteEdge.FromID, remoteEdge.ToID, remoteEdge.Relation)
		if _, exists := localEdgeMap[key]; !exists {
			if err := s.db.ImportEdges([]*MemoryEdge{remoteEdge}); err != nil {
				logger.Warn("Failed to import edge",
					"from", remoteEdge.FromID,
					"to", remoteEdge.ToID,
					"error", err)
			}
		}
	}

	return nil
}

// edgeKey generates a unique key for an edge
func edgeKey(fromID, toID, relation string) string {
	return fmt.Sprintf("%s:%s:%s", fromID, toID, relation)
}

// generateCursor generates a cursor based on memories
func generateCursor(memories []*Memory) string {
	if len(memories) == 0 {
		return ""
	}

	// Sort by created_at
	sorted := make([]*Memory, len(memories))
	copy(sorted, memories)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})

	// Return the ID of the most recent memory
	return sorted[len(sorted)-1].ID
}

// getSyncFilePath returns the path to the sync state file
func (s *Syncer) getSyncFilePath(remote string) string {
	homeDir, _ := os.UserHomeDir()
	syncDir := filepath.Join(homeDir, ".odin", "sync")
	os.MkdirAll(syncDir, 0755)
	return filepath.Join(syncDir, "mimir_sync.json")
}

// LoadSyncState loads the synchronization state
func (s *Syncer) LoadSyncState(remote string) (*SyncState, error) {
	syncPath := s.getSyncFilePath(remote)

	data, err := os.ReadFile(syncPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read sync state: %w", err)
	}

	var state SyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to deserialize sync state: %w", err)
	}

	return &state, nil
}

// SaveSyncState saves the synchronization state
func (s *Syncer) SaveSyncState(state *SyncState) error {
	syncPath := s.getSyncFilePath(state.RemoteURL)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize sync state: %w", err)
	}

	if err := os.WriteFile(syncPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write sync state: %w", err)
	}

	return nil
}

// HasConflicts checks if there are any sync conflicts
func (s *Syncer) HasConflicts(local, remote []*Memory) bool {
	localMap := make(map[string]*Memory)
	for _, m := range local {
		localMap[m.ID] = m
	}

	for _, r := range remote {
		if l, exists := localMap[r.ID]; exists {
			// Same ID exists locally and remotely
			// Check if they're different
			if l.Content != r.Content && !l.UpdatedAt.Equal(r.UpdatedAt) {
				// Potential conflict
				return true
			}
		}
	}

	return false
}

// ResolveConflict resolves a conflict by choosing the newer version
func (s *Syncer) ResolveConflict(local, remote *Memory) *Memory {
	if local.UpdatedAt.After(remote.UpdatedAt) {
		return local
	}
	return remote
}

// GetSyncStatus returns the current sync status
func (s *Syncer) GetSyncStatus(remote string) (string, error) {
	state, err := s.LoadSyncState(remote)
	if err != nil {
		return "", err
	}

	if state == nil {
		return "never_synced", nil
	}

	if time.Since(state.LastSync) < 5*time.Minute {
		return "synced", nil
	}

	return "stale", nil
}

// ExportMemories exports all memories to a JSON file
func (s *Syncer) ExportMemories(path string) error {
	memories, err := s.db.GetAllMemories()
	if err != nil {
		return fmt.Errorf("failed to get memories: %w", err)
	}

	edges, err := s.db.GetAllEdges()
	if err != nil {
		return fmt.Errorf("failed to get edges: %w", err)
	}

	data := RemoteMemories{
		Memories: memories,
		Edges:    edges,
		State: SyncState{
			LastSync: time.Now(),
		},
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize: %w", err)
	}

	return os.WriteFile(path, jsonData, 0644)
}

// ImportMemories imports memories from a JSON file
func (s *Syncer) ImportMemoriesFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var remoteData RemoteMemories
	if err := json.Unmarshal(data, &remoteData); err != nil {
		return fmt.Errorf("failed to deserialize: %w", err)
	}

	return s.merge(remoteData)
}
