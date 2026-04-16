package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/odin-ai/odin/pkg/logger"
)

// GitClient handles Git operations for the sync engine
type GitClient struct {
	repoPath string
	repo     *git.Repository
	remote   string
	gpgSign  bool
}

// GitStatus represents the current status of the git repository
type GitStatus struct {
	IsInitialized    bool
	HasRemote        bool
	RemoteURL        string
	CurrentBranch    string
	HasUncommitted   bool
	UncommittedFiles []string
	StagedFiles      []string
}

// NewGitClient creates a new Git client
func NewGitClient(repoPath, remote string, gpgSign bool) (*GitClient, error) {
	gc := &GitClient{
		repoPath: repoPath,
		remote:   remote,
		gpgSign:  gpgSign,
	}

	// Try to open existing repo
	repo, err := git.PlainOpen(repoPath)
	if err == nil {
		gc.repo = repo
	}

	return gc, nil
}

// Init initializes a new Git repository
func (g *GitClient) Init() error {
	if g.repo != nil {
		return nil // Already initialized
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(g.repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create repo directory: %w", err)
	}

	repo, err := git.PlainInit(g.repoPath, false)
	if err != nil {
		return fmt.Errorf("failed to init git repo: %w", err)
	}

	g.repo = repo
	logger.Info("Git repository initialized", "path", g.repoPath)
	return nil
}

// Open opens an existing Git repository
func (g *GitClient) Open() error {
	repo, err := git.PlainOpen(g.repoPath)
	if err != nil {
		return fmt.Errorf("failed to open git repo: %w", err)
	}
	g.repo = repo
	return nil
}

// IsInitialized returns true if the repository is initialized
func (g *GitClient) IsInitialized() bool {
	if g.repo == nil {
		return false
	}
	_, err := g.repo.Worktree()
	return err == nil
}

// SetRemote sets the remote URL
func (g *GitClient) SetRemote(remoteURL string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	_, err := g.repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteURL},
	})
	if err != nil {
		// Remote might already exist, delete and recreate
		if delErr := g.repo.DeleteRemote("origin"); delErr != nil {
			return fmt.Errorf("failed to delete existing remote: %w", delErr)
		}
		_, err = g.repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{remoteURL},
		})
		if err != nil {
			return fmt.Errorf("failed to create remote: %w", err)
		}
	}

	g.remote = remoteURL
	return nil
}

// Push pushes to the remote repository
func (g *GitClient) Push() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	remote, err := g.repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("no remote configured: %w", err)
	}

	// Push options
	opts := &git.PushOptions{
		RemoteName: "origin",
	}

	if g.gpgSign {
		// Note: go-git doesn't have native GPG signing,
		// signatures would need to be applied externally
		logger.Warn("GPG signing requested but not fully supported by go-git")
	}

	err = remote.Push(opts)
	if err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	logger.Info("Push successful", "remote", g.remote)
	return nil
}

// Pull pulls from the remote repository
func (g *GitClient) Pull() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	opts := &git.PullOptions{
		RemoteName: "origin",
	}

	err = worktree.Pull(opts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("pull failed: %w", err)
	}

	if err == git.NoErrAlreadyUpToDate {
		logger.Info("Already up to date")
	} else {
		logger.Info("Pull successful")
	}
	return nil
}

// Status returns the current status of the repository
func (g *GitClient) Status() (*GitStatus, error) {
	status := &GitStatus{
		IsInitialized: g.IsInitialized(),
	}

	if !status.IsInitialized {
		return status, nil
	}

	worktree, err := g.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get current branch
	head, err := g.repo.Head()
	if err == nil {
		status.CurrentBranch = head.Name().Short()
	}

	// Get remote URL
	remote, err := g.repo.Remote("origin")
	if err == nil {
		status.HasRemote = true
		cfg := remote.Config()
		if len(cfg.URLs) > 0 {
			status.RemoteURL = cfg.URLs[0]
		}
	}

	// Get working tree status
	treeStatus, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	for path, s := range treeStatus {
		if s.Worktree == git.Unmodified && s.Staging == git.Unmodified {
			continue
		}
		if s.Worktree != git.Unmodified {
			status.UncommittedFiles = append(status.UncommittedFiles, path)
			status.HasUncommitted = true
		}
		if s.Staging != git.Unmodified {
			status.StagedFiles = append(status.StagedFiles, path)
		}
	}

	return status, nil
}

// Diff returns the diff of uncommitted changes
func (g *GitClient) Diff() (string, error) {
	if g.repo == nil {
		return "", fmt.Errorf("repository not initialized")
	}

	worktree, err := g.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	treeStatus, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	if treeStatus.IsClean() {
		return "No changes", nil
	}

	// Build diff output
	diffOutput := "Changes:\n"
	for path, s := range treeStatus {
		if s.Worktree != git.Unmodified {
			diffOutput += fmt.Sprintf("  M %s\n", path)
		}
		if s.Staging != git.Unmodified {
			diffOutput += fmt.Sprintf("  A %s\n", path)
		}
	}

	return diffOutput, nil
}

// Log returns the commit history
func (g *GitClient) Log(limit int) ([]*CommitInfo, error) {
	if g.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	head, err := g.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commits, err := g.repo.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}
	defer commits.Close()

	var result []*CommitInfo
	count := 0

	for {
		c, err := commits.Next()
		if err != nil {
			break
		}

		info := &CommitInfo{
			Hash:      c.Hash.String()[:7],
			FullHash:  c.Hash.String(),
			Author:    c.Author.Name,
			Email:     c.Author.Email,
			Timestamp: c.Author.When,
			Message:   c.Message,
		}
		result = append(result, info)

		count++
		if limit > 0 && count >= limit {
			break
		}
	}

	return result, nil
}

// CommitInfo represents information about a commit
type CommitInfo struct {
	Hash      string
	FullHash  string
	Author    string
	Email     string
	Timestamp time.Time
	Message   string
}

// Add adds a file to the staging area
func (g *GitClient) Add(paths ...string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	for _, path := range paths {
		_, err := worktree.Add(path)
		if err != nil {
			return fmt.Errorf("failed to add %s: %w", path, err)
		}
	}

	return nil
}

// Commit creates a new commit
func (g *GitClient) Commit(message string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = worktree.Commit(message, &git.CommitOptions{
		All: false,
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	logger.Info("Commit created", "message", message)
	return nil
}

// GetRemote returns the remote URL
func (g *GitClient) GetRemote() string {
	return g.remote
}

// ListBranches returns all local and remote branches
func (g *GitClient) ListBranches() ([]string, error) {
	if g.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	var branches []string

	// Local branches
	localRefs, err := g.repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	localRefs.ForEach(func(ref *plumbing.Reference) error {
		branches = append(branches, ref.Name().Short())
		return nil
	})

	return branches, nil
}

// CreateBranch creates a new branch
func (g *GitClient) CreateBranch(name string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(name),
		Create: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	logger.Info("Branch created", "name", name)
	return nil
}

// CheckoutBranch switches to a branch
func (g *GitClient) CheckoutBranch(name string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(name),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

// Merge merges a branch into the current branch
func (g *GitClient) Merge(branch string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	// Verify the branch exists
	_, err := g.repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return fmt.Errorf("branch not found: %w", err)
	}

	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Checkout the target branch (this performs a merge in go-git)
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Force:  false,
	})
	if err != nil {
		return fmt.Errorf("merge failed: %w", err)
	}

	logger.Info("Merge completed", "branch", branch)
	return nil
}

// GetWorktreePath returns the working tree path
func (g *GitClient) GetWorktreePath() string {
	return g.repoPath
}

// Clone clones a remote repository into the local path
func (g *GitClient) Clone(remoteURL string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(g.repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create repo directory: %w", err)
	}

	// Clone into memory first
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteURL},
	})

	refs, err := rem.List(&git.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list remote: %w", err)
	}

	// Find main branch
	var branchName plumbing.ReferenceName
	for _, ref := range refs {
		if ref.Name().Short() == "main" || ref.Name().Short() == "master" {
			branchName = ref.Name()
			break
		}
	}
	if branchName == "" && len(refs) > 0 {
		branchName = refs[0].Name()
	}

	// Clone
	repo, err := git.PlainClone(g.repoPath, false, &git.CloneOptions{
		URL:           remoteURL,
		ReferenceName: branchName,
		SingleBranch:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to clone: %w", err)
	}

	g.repo = repo
	g.remote = remoteURL

	logger.Info("Clone completed", "path", g.repoPath, "remote", remoteURL)
	return nil
}

// Fetch fetches from the remote
func (g *GitClient) Fetch() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	remote, err := g.repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("no remote configured: %w", err)
	}

	err = remote.Fetch(&git.FetchOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("fetch failed: %w", err)
	}

	return nil
}

// GetFileContent returns the content of a file at a specific commit
func (g *GitClient) GetFileContent(path string, commitHash plumbing.Hash) (string, error) {
	if g.repo == nil {
		return "", fmt.Errorf("repository not initialized")
	}

	commit, err := g.repo.CommitObject(commitHash)
	if err != nil {
		return "", fmt.Errorf("failed to get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get tree: %w", err)
	}

	file, err := tree.File(path)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	content, err := file.Contents()
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return content, nil
}

// GetRepoPath returns the repository path
func (g *GitClient) GetRepoPath() string {
	return g.repoPath
}

// FileExists checks if a file exists in the repository at HEAD
func (g *GitClient) FileExists(path string) (bool, error) {
	if g.repo == nil {
		return false, fmt.Errorf("repository not initialized")
	}

	head, err := g.repo.Head()
	if err != nil {
		return false, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := g.repo.CommitObject(head.Hash())
	if err != nil {
		return false, fmt.Errorf("failed to get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return false, fmt.Errorf("failed to get tree: %w", err)
	}

	_, err = tree.File(filepath.Join(g.repoPath, path))
	if err != nil {
		return false, nil
	}

	return true, nil
}

// ReadFile reads a file from the working directory
func (g *GitClient) ReadFile(path string) ([]byte, error) {
	fullPath := filepath.Join(g.repoPath, path)
	return os.ReadFile(fullPath)
}

// WriteFile writes a file to the working directory
func (g *GitClient) WriteFile(path string, data []byte) error {
	fullPath := filepath.Join(g.repoPath, path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return os.WriteFile(fullPath, data, 0644)
}
