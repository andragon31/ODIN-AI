package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/odin-ai/odin/pkg/logger"
)

// Bifrost is the sync engine that manages Git-based configuration sync
type Bifrost struct {
	repoPath string
	remote   string
	gpgSign  bool
	crdt     *CRDTEngine
	git      *GitClient
	nodeID   string
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Success   bool
	Pushed    bool
	Pulled    bool
	Conflicts []Conflict
	Commits   int
	Error     error
}

// Conflict represents a sync conflict
type Conflict struct {
	Path       string
	Local      string
	Remote     string
	Resolved   bool
	Resolution string // "local", "remote", "merged"
}

// NewBifrost creates a new Bifrost sync engine
func NewBifrost(repoPath, remote string, gpgSign bool) (*Bifrost, error) {
	nodeID := uuid.New().String()[:8]
	crdt := NewCRDTEngine(nodeID)

	git, err := NewGitClient(repoPath, remote, gpgSign)
	if err != nil {
		return nil, err
	}

	b := &Bifrost{
		repoPath: repoPath,
		remote:   remote,
		gpgSign:  gpgSign,
		crdt:     crdt,
		git:      git,
		nodeID:   nodeID,
	}

	return b, nil
}

// Init initializes the sync repository
func (b *Bifrost) Init() error {
	if err := b.git.Init(); err != nil {
		return fmt.Errorf("failed to init git repo: %w", err)
	}

	if b.remote != "" {
		if err := b.git.SetRemote(b.remote); err != nil {
			logger.Warn("Failed to set remote", "error", err)
		}
	}

	logger.Info("Bifrost initialized", "path", b.repoPath, "remote", b.remote)
	return nil
}

// IsInitialized returns true if the sync repository is initialized
func (b *Bifrost) IsInitialized() bool {
	return b.git.IsInitialized()
}

// SetRemote sets the remote URL
func (b *Bifrost) SetRemote(remote string) error {
	if !b.git.IsInitialized() {
		return fmt.Errorf("repository not initialized")
	}

	if err := b.git.SetRemote(remote); err != nil {
		return fmt.Errorf("failed to set remote: %w", err)
	}

	b.remote = remote
	return nil
}

// Push pushes changes to the remote
func (b *Bifrost) Push() error {
	if !b.git.IsInitialized() {
		return fmt.Errorf("repository not initialized")
	}

	if b.remote == "" {
		return fmt.Errorf("no remote configured")
	}

	if err := b.git.Push(); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	return nil
}

// Pull pulls changes from the remote
func (b *Bifrost) Pull() (*SyncResult, error) {
	result := &SyncResult{Success: true}

	if !b.git.IsInitialized() {
		return nil, fmt.Errorf("repository not initialized")
	}

	if b.remote == "" {
		return nil, fmt.Errorf("no remote configured")
	}

	// Get status before pull
	beforeStatus, err := b.git.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Fetch first to get remote changes
	if err := b.git.Fetch(); err != nil {
		logger.Warn("Fetch failed", "error", err)
	}

	// Pull
	if err := b.git.Pull(); err != nil {
		return nil, fmt.Errorf("pull failed: %w", err)
	}
	result.Pulled = true

	// Get status after pull
	afterStatus, err := b.git.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Detect conflicts by comparing before and after
	result.Conflicts = b.detectConflicts(beforeStatus, afterStatus)

	// Count new commits
	commits, err := b.git.Log(10)
	if err == nil {
		result.Commits = len(commits)
	}

	result.Success = len(result.Conflicts) == 0
	return result, nil
}

// detectConflicts identifies conflicts between local and remote changes
func (b *Bifrost) detectConflicts(before, after *GitStatus) []Conflict {
	var conflicts []Conflict

	// Files that were modified in both before and after are potential conflicts
	beforeMap := make(map[string]bool)
	for _, f := range before.UncommittedFiles {
		beforeMap[f] = true
	}

	for _, f := range after.UncommittedFiles {
		if beforeMap[f] {
			conflicts = append(conflicts, Conflict{
				Path:     f,
				Resolved: false,
			})
		}
	}

	return conflicts
}

// Status returns the current sync status
func (b *Bifrost) Status() (*BifrostStatus, error) {
	status := &BifrostStatus{
		Initialized: b.git.IsInitialized(),
		RepoPath:    b.repoPath,
		Remote:      b.remote,
		GPGSign:     b.gpgSign,
	}

	if !status.Initialized {
		return status, nil
	}

	gitStatus, err := b.git.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}

	status.HasRemote = gitStatus.HasRemote
	status.CurrentBranch = gitStatus.CurrentBranch
	status.HasUncommitted = gitStatus.HasUncommitted
	status.UncommittedFiles = gitStatus.UncommittedFiles
	status.StagedFiles = gitStatus.StagedFiles

	return status, nil
}

// BifrostStatus represents the status of the Bifrost sync engine
type BifrostStatus struct {
	Initialized      bool
	RepoPath         string
	Remote           string
	HasRemote        bool
	CurrentBranch    string
	HasUncommitted   bool
	UncommittedFiles []string
	StagedFiles      []string
	GPGSign          bool
}

// Diff returns the diff of uncommitted changes
func (b *Bifrost) Diff() (string, error) {
	if !b.git.IsInitialized() {
		return "", fmt.Errorf("repository not initialized")
	}

	return b.git.Diff()
}

// Log returns the commit history
func (b *Bifrost) Log(limit int) ([]*CommitInfo, error) {
	if !b.git.IsInitialized() {
		return nil, fmt.Errorf("repository not initialized")
	}

	return b.git.Log(limit)
}

// BranchList returns all branches
func (b *Bifrost) BranchList() ([]*BranchInfo, error) {
	if !b.git.IsInitialized() {
		return nil, fmt.Errorf("repository not initialized")
	}

	return b.git.BranchList()
}

// BranchCreate creates a new branch
func (b *Bifrost) BranchCreate(name string) error {
	if !b.git.IsInitialized() {
		return fmt.Errorf("repository not initialized")
	}

	return b.git.BranchCreate(name)
}

// Merge merges a branch
func (b *Bifrost) Merge(branch string) error {
	if !b.git.IsInitialized() {
		return fmt.Errorf("repository not initialized")
	}

	return b.git.Merge(branch)
}

// SetGPGSign enables or disables GPG signing
func (b *Bifrost) SetGPGSign(enabled bool) {
	b.gpgSign = enabled
}

// GetGPGSign returns whether GPG signing is enabled
func (b *Bifrost) GetGPGSign() bool {
	return b.gpgSign
}

// GetRepoPath returns the repository path
func (b *Bifrost) GetRepoPath() string {
	return b.repoPath
}

// ReadFile reads a file from the sync repository
func (b *Bifrost) ReadFile(path string) ([]byte, error) {
	if !b.git.IsInitialized() {
		return nil, fmt.Errorf("repository not initialized")
	}

	return b.git.ReadFile(path)
}

// WriteFile writes a file to the sync repository
func (b *Bifrost) WriteFile(path string, data []byte) error {
	if !b.git.IsInitialized() {
		return fmt.Errorf("repository not initialized")
	}

	return b.git.WriteFile(path, data)
}

// AddAndCommit adds files and creates a commit
func (b *Bifrost) AddAndCommit(message string, paths ...string) error {
	if !b.git.IsInitialized() {
		return fmt.Errorf("repository not initialized")
	}

	if err := b.git.Add(paths...); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	return b.git.Commit(message)
}

// Sync performs a full sync (pull then push)
func (b *Bifrost) Sync() (*SyncResult, error) {
	result := &SyncResult{Success: true}

	// Pull first
	pullResult, err := b.Pull()
	if err != nil {
		return nil, fmt.Errorf("pull failed: %w", err)
	}
	result.Pulled = pullResult.Pulled
	result.Conflicts = append(result.Conflicts, pullResult.Conflicts...)

	// Then push
	if err := b.Push(); err != nil {
		result.Success = false
		result.Error = err
		return result, nil
	}
	result.Pushed = true

	result.Success = len(result.Conflicts) == 0
	return result, nil
}

// GetCRDTEngine returns the CRDT engine for conflict resolution
func (b *Bifrost) GetCRDTEngine() *CRDTEngine {
	return b.crdt
}

// ResolveConflict resolves a conflict using LWW strategy
func (b *Bifrost) ResolveConflict(conflict *Conflict, useLocal bool) error {
	// Use CRDT's LWW strategy
	// For simplicity, just mark as resolved
	conflict.Resolved = true
	if useLocal {
		conflict.Resolution = "local"
	} else {
		conflict.Resolution = "remote"
	}
	return nil
}

// DefaultRepoPath returns the default sync repository path
func DefaultRepoPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".odin", "config")
}

// DefaultRemote returns the default remote based on environment
func DefaultRemote() string {
	// Could be set from environment or config
	return ""
}

// CreateDefault creates a Bifrost instance with default settings
func CreateDefault() (*Bifrost, error) {
	repoPath := DefaultRepoPath()
	remote := DefaultRemote()
	return NewBifrost(repoPath, remote, false)
}

// LastSyncTime returns the time of the last sync operation
// This is tracked via commit timestamps
func (b *Bifrost) LastSyncTime() (*time.Time, error) {
	if !b.git.IsInitialized() {
		return nil, nil
	}

	commits, err := b.git.Log(1)
	if err != nil || len(commits) == 0 {
		return nil, err
	}

	return &commits[0].Timestamp, nil
}

// GetNodeID returns the unique node ID for this Bifrost instance
func (b *Bifrost) GetNodeID() string {
	return b.nodeID
}
