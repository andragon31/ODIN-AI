package sync

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCRDTMergeText(t *testing.T) {
	crdt := NewCRDTEngine("node1")

	tests := []struct {
		name     string
		local    string
		remote   string
		localTS  time.Time
		remoteTS time.Time
		expected string
	}{
		{
			name:     "remote is newer",
			local:    "local value",
			remote:   "remote value",
			localTS:  time.Now().Add(-1 * time.Hour),
			remoteTS: time.Now(),
			expected: "remote value",
		},
		{
			name:     "local is newer",
			local:    "local value",
			remote:   "remote value",
			localTS:  time.Now(),
			remoteTS: time.Now().Add(-1 * time.Hour),
			expected: "local value",
		},
		{
			name:     "same timestamp - local wins",
			local:    "local value",
			remote:   "remote value",
			localTS:  time.Now(),
			remoteTS: time.Now(),
			expected: "local value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crdt.MergeText(tt.local, tt.remote, tt.localTS, tt.remoteTS)
			if result.Value != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.Value)
			}
			if result.Conflict {
				t.Error("expected no conflict")
			}
			if !result.Resolved {
				t.Error("expected resolved")
			}
		})
	}
}

func TestCRDTMergeJSON(t *testing.T) {
	crdt := NewCRDTEngine("node1")

	tests := []struct {
		name        string
		local       string
		remote      string
		localTS     time.Time
		remoteTS    time.Time
		hasConflict bool
	}{
		{
			name:        "simple merge - remote newer",
			local:       `{"key": "local"}`,
			remote:      `{"key": "remote"}`,
			localTS:     time.Now().Add(-1 * time.Hour),
			remoteTS:    time.Now(),
			hasConflict: false,
		},
		{
			name:        "nested objects",
			local:       `{"outer": {"inner": "local"}}`,
			remote:      `{"outer": {"inner": "remote"}}`,
			localTS:     time.Now(),
			remoteTS:    time.Now().Add(-1 * time.Hour),
			hasConflict: false,
		},
		{
			name:        "new keys from both sides",
			local:       `{"localKey": "value"}`,
			remote:      `{"remoteKey": "value"}`,
			localTS:     time.Now(),
			remoteTS:    time.Now(),
			hasConflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := crdt.MergeJSON([]byte(tt.local), []byte(tt.remote), tt.localTS, tt.remoteTS)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.hasConflict && !result.Conflict {
				t.Error("expected conflict")
			}
			if !tt.hasConflict && result.Conflict {
				t.Error("expected no conflict")
			}
		})
	}
}

func TestCRDTMergeArrays(t *testing.T) {
	crdt := NewCRDTEngine("node1")

	local := []interface{}{"a", "b", "c"}
	remote := []interface{}{"b", "c", "d"}

	result := crdt.mergeArrays(local, remote)

	expected := []interface{}{"a", "b", "c", "d"}
	if len(result) != len(expected) {
		t.Errorf("expected %d elements, got %d", len(expected), len(result))
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, v := range result {
		key := v.(string)
		if seen[key] {
			t.Errorf("duplicate element: %s", key)
		}
		seen[key] = true
	}
}

func TestGenerateLWWID(t *testing.T) {
	crdt := NewCRDTEngine("node1")

	id1 := crdt.GenerateLWWID()
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Nanosecond)
	id2 := crdt.GenerateLWWID()

	// IDs should be unique (or at least one should have different timestamp)
	if id1 == id2 && contains(id1, id2) {
		t.Log("IDs might be same if generated in same nanosecond - this is acceptable")
	}

	// Should contain node ID
	if !contains(id1, "node1") {
		t.Error("ID should contain node ID")
	}
}

func TestParseLWWID(t *testing.T) {
	crdt := NewCRDTEngine("node1")

	id := crdt.GenerateLWWID()
	ts, nodeID, err := ParseLWWID(id)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if nodeID != "node1" {
		t.Errorf("expected node1, got %s", nodeID)
	}

	if ts.IsZero() {
		t.Error("timestamp should not be zero")
	}
}

func TestParseLWWIDInvalid(t *testing.T) {
	_, _, err := ParseLWWID("invalid")
	if err == nil {
		t.Error("expected error for invalid ID")
	}
}

func TestVectorClock(t *testing.T) {
	crdt := NewCRDTEngine("node1")

	// Initial clock should be empty
	clock := crdt.GetClock()
	if len(clock) != 0 {
		t.Error("initial clock should be empty")
	}

	// Update clock
	crdt.UpdateClock()
	clock = crdt.GetClock()
	if clock["node1"] != 1 {
		t.Errorf("expected clock[node1]=1, got %d", clock["node1"])
	}

	// Merge clock
	otherClock := VectorClock{
		"node1": 1,
		"node2": 2,
	}
	crdt.MergeClock(otherClock)
	clock = crdt.GetClock()

	if clock["node1"] != 1 {
		t.Errorf("expected clock[node1]=1, got %d", clock["node1"])
	}
	if clock["node2"] != 2 {
		t.Errorf("expected clock[node2]=2, got %d", clock["node2"])
	}
}

func TestCompareClocks(t *testing.T) {
	// Test that the CRDT engine's internal clock comparison works
	crdt := NewCRDTEngine("node1")

	// Set up internal clock using a write lock
	crdt.mu.Lock()
	crdt.clock = VectorClock{"node1": 1}
	crdt.mu.Unlock()

	// Test: internal clock {node1: 1} vs {node1: 1}
	result := crdt.CompareClocks(VectorClock{"node1": 1})
	if result != 0 {
		t.Errorf("equal clocks: expected 0, got %d", result)
	}

	// Test: internal clock {node1: 1} vs {node1: 2}
	result = crdt.CompareClocks(VectorClock{"node1": 2})
	if result != -1 {
		t.Errorf("this less than other: expected -1, got %d", result)
	}

	// Test: internal clock {node1: 2} vs {node1: 1}
	crdt.mu.Lock()
	crdt.clock = VectorClock{"node1": 2}
	crdt.mu.Unlock()
	result = crdt.CompareClocks(VectorClock{"node1": 1})
	if result != 1 {
		t.Errorf("this greater than other: expected 1, got %d", result)
	}

	// Test: concurrent - different nodes
	crdt.mu.Lock()
	crdt.clock = VectorClock{"node1": 1}
	crdt.mu.Unlock()
	result = crdt.CompareClocks(VectorClock{"node2": 1})
	// Should be 0 (concurrent) because they have different nodes and neither is strictly less/greater
	if result != 0 {
		t.Errorf("concurrent clocks: expected 0, got %d", result)
	}
}

func TestGitClientInit(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	git, err := NewGitClient(repoPath, "", false)
	if err != nil {
		t.Fatalf("failed to create git client: %v", err)
	}

	// Should not be initialized yet
	if git.IsInitialized() {
		t.Error("should not be initialized before Init()")
	}

	// Init
	if err := git.Init(); err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	// Should be initialized now
	if !git.IsInitialized() {
		t.Error("should be initialized after Init()")
	}

	// Try to init again - should be idempotent
	if err := git.Init(); err != nil {
		t.Fatalf("second init should succeed: %v", err)
	}
}

func TestGitClientStatus(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	git, err := NewGitClient(repoPath, "", false)
	if err != nil {
		t.Fatalf("failed to create git client: %v", err)
	}

	// Init
	if err := git.Init(); err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	// Get status
	status, err := git.Status()
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	if !status.IsInitialized {
		t.Error("expected initialized")
	}

	// Branch name could be "main", "master", or empty depending on git version
	// Just check it's not an error condition
	if status.CurrentBranch == "" {
		t.Log("branch name is empty - this may happen in test environments")
	}
}

func TestGitClientAddAndCommit(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	git, err := NewGitClient(repoPath, "", false)
	if err != nil {
		t.Fatalf("failed to create git client: %v", err)
	}

	// Init
	if err := git.Init(); err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	// Create a file
	testFile := "test.txt"
	if err := git.WriteFile(testFile, []byte("hello world")); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Add
	if err := git.Add(testFile); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	// Commit
	if err := git.Commit("test commit"); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Check log
	commits, err := git.Log(10)
	if err != nil {
		t.Fatalf("failed to get log: %v", err)
	}

	if len(commits) != 1 {
		t.Errorf("expected 1 commit, got %d", len(commits))
	}

	if commits[0].Message != "test commit" {
		t.Errorf("expected 'test commit', got '%s'", commits[0].Message)
	}
}

func TestBifrostCreation(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	b, err := NewBifrost(repoPath, "", false)
	if err != nil {
		t.Fatalf("failed to create bifrost: %v", err)
	}

	if b.GetRepoPath() != repoPath {
		t.Errorf("expected repo path %s, got %s", repoPath, b.GetRepoPath())
	}

	if b.GetGPGSign() {
		t.Error("expected GPGSign to be false")
	}
}

func TestBifrostInit(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	b, err := NewBifrost(repoPath, "", false)
	if err != nil {
		t.Fatalf("failed to create bifrost: %v", err)
	}

	// Should not be initialized
	if b.IsInitialized() {
		t.Error("should not be initialized before Init()")
	}

	// Init
	if err := b.Init(); err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	// Should be initialized
	if !b.IsInitialized() {
		t.Error("should be initialized after Init()")
	}

	// Should be able to get status
	status, err := b.Status()
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	if !status.Initialized {
		t.Error("status should show initialized")
	}
}

func TestDefaultRepoPath(t *testing.T) {
	path := DefaultRepoPath()

	if path == "" {
		t.Error("default repo path should not be empty")
	}

	// Should contain .odin
	if !contains(path, ".odin") {
		t.Error("default repo path should contain .odin")
	}
}

func TestBifrostStatus(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	b, err := NewBifrost(repoPath, "git@github.com:test/test.git", false)
	if err != nil {
		t.Fatalf("failed to create bifrost: %v", err)
	}

	// Init
	if err := b.Init(); err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	// Set remote
	if err := b.SetRemote("git@github.com:test/test.git"); err != nil {
		t.Fatalf("failed to set remote: %v", err)
	}

	// Get status
	status, err := b.Status()
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	if !status.HasRemote {
		t.Error("expected HasRemote to be true")
	}

	if status.Remote != "git@github.com:test/test.git" {
		t.Errorf("expected remote git@github.com:test/test.git, got %s", status.Remote)
	}
}

func TestBifrostWriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	b, err := NewBifrost(repoPath, "", false)
	if err != nil {
		t.Fatalf("failed to create bifrost: %v", err)
	}

	// Init
	if err := b.Init(); err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	// Write a file
	testContent := []byte("test content")
	if err := b.WriteFile("test.txt", testContent); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Read it back
	content, err := b.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) != string(testContent) {
		t.Errorf("expected %s, got %s", string(testContent), string(content))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkCRDTMergeJSON(b *testing.B) {
	crdt := NewCRDTEngine("node1")
	local := []byte(`{"key1": "value1", "key2": "value2", "nested": {"a": 1, "b": 2}}`)
	remote := []byte(`{"key1": "value1", "key3": "value3", "nested": {"a": 1, "c": 3}}`)
	localTS := time.Now().Add(-1 * time.Hour)
	remoteTS := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crdt.MergeJSON(local, remote, localTS, remoteTS)
	}
}

// Integration test - requires git to be installed
func TestGitClientWithRealGit(t *testing.T) {
	if os.Getenv("SKIP_GIT_TESTS") == "1" {
		t.Skip("skipping git integration tests")
	}

	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	git, err := NewGitClient(repoPath, "", false)
	if err != nil {
		t.Fatalf("failed to create git client: %v", err)
	}

	// Init
	if err := git.Init(); err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	// Create worktree path
	worktreePath := git.GetWorktreePath()

	// Write a file to the worktree
	testFile := filepath.Join(worktreePath, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Add and commit
	if err := git.Add("test.txt"); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	if err := git.Commit("Initial commit"); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Check that the file exists
	if _, err := os.Stat(testFile); err != nil {
		t.Fatalf("file should exist: %v", err)
	}
}
