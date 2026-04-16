package sync

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/odin-ai/odin/pkg/logger"
)

// BranchInfo represents information about a branch
type BranchInfo struct {
	Name   string
	Hash   string
	IsHead bool
}

// ListBranches returns all local branches
func (g *GitClient) BranchList() ([]*BranchInfo, error) {
	if g.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	head, err := g.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	currentBranch := head.Name().Short()
	var branches []*BranchInfo

	// Iterate over local branches
	branchIter, err := g.repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to iterate branches: %w", err)
	}

	err = branchIter.ForEach(func(ref *plumbing.Reference) error {
		info := &BranchInfo{
			Name:   ref.Name().Short(),
			Hash:   ref.Hash().String()[:7],
			IsHead: ref.Name().Short() == currentBranch,
		}
		branches = append(branches, info)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	return branches, nil
}

// BranchCreate creates a new branch
func (g *GitClient) BranchCreate(name string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	// Check if branch already exists
	refs, err := g.repo.Branches()
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	var exists bool
	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().Short() == name {
			exists = true
		}
		return nil
	})

	if exists {
		return fmt.Errorf("branch '%s' already exists", name)
	}

	// Create the branch
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

// BranchDelete deletes a branch
func (g *GitClient) BranchDelete(name string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	// Get current branch
	head, err := g.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	currentBranch := head.Name().Short()
	if currentBranch == name {
		return fmt.Errorf("cannot delete current branch '%s'", name)
	}

	err = g.repo.DeleteBranch(name)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	logger.Info("Branch deleted", "name", name)
	return nil
}

// BranchMerge merges a source branch into the current branch
func (g *GitClient) BranchMerge(source string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	// Verify the source branch exists
	_, err := g.repo.Reference(plumbing.NewBranchReferenceName(source), true)
	if err != nil {
		return fmt.Errorf("source branch '%s' not found: %w", source, err)
	}

	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Checkout the source branch to merge it
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(source),
		Force:  false,
	})
	if err != nil {
		return fmt.Errorf("merge failed: %w", err)
	}

	logger.Info("Branch merged", "source", source)
	return nil
}

// GetCurrentBranch returns the current branch name
func (g *GitClient) GetCurrentBranch() (string, error) {
	if g.repo == nil {
		return "", fmt.Errorf("repository not initialized")
	}

	head, err := g.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return head.Name().Short(), nil
}

// HasDiverged checks if the current branch has diverged from another branch
func (g *GitClient) HasDiverged(branch string) (bool, error) {
	if g.repo == nil {
		return false, fmt.Errorf("repository not initialized")
	}

	// Get the other branch reference
	ref, err := g.repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return false, fmt.Errorf("branch not found: %w", err)
	}

	// Get current branch reference
	head, err := g.repo.Head()
	if err != nil {
		return false, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get the common ancestor
	mergeBase, err := g.repoMergeBase(head.Hash(), ref.Hash())
	if err != nil {
		return false, fmt.Errorf("failed to find merge base: %w", err)
	}

	// If merge base equals the other branch, current is behind
	// If merge base equals current head, current is ahead
	// Otherwise, they have diverged
	hasDiverged := mergeBase != head.Hash() && mergeBase != ref.Hash()
	return hasDiverged, nil
}

// GetAheadBehind returns the number of commits ahead and behind
func (g *GitClient) GetAheadBehind(branch string) (ahead, behind int, err error) {
	if g.repo == nil {
		return 0, 0, fmt.Errorf("repository not initialized")
	}

	// Get the other branch reference
	ref, err := g.repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return 0, 0, fmt.Errorf("branch not found: %w", err)
	}

	head, err := g.repo.Head()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Count commits ahead (from other to current)
	ahead, err = g.countCommitsRange(ref.Hash(), head.Hash())
	if err != nil {
		return 0, 0, err
	}

	// Count commits behind (from current to other)
	behind, err = g.countCommitsRange(head.Hash(), ref.Hash())
	if err != nil {
		return 0, 0, err
	}

	return ahead, behind, nil
}

// countCommitsRange counts commits from start to end (excluding start, including end)
func (g *GitClient) countCommitsRange(start, end plumbing.Hash) (int, error) {
	count := 0

	commitIter, err := g.repo.Log(&git.LogOptions{
		From: end,
	})
	if err != nil {
		return 0, err
	}
	defer commitIter.Close()

	for {
		c, err := commitIter.Next()
		if err != nil {
			break
		}

		if c.Hash == start {
			break
		}
		count++
	}

	return count, nil
}

// repoMergeBase finds the common ancestor of two commits
// This is a simplified implementation using commit ancestry traversal
func (g *GitClient) repoMergeBase(h1, h2 plumbing.Hash) (plumbing.Hash, error) {
	// Get the first commit
	c1, err := g.repo.CommitObject(h1)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get commit1: %w", err)
	}

	// Build a set of all ancestors of h1
	ancestors := make(map[plumbing.Hash]bool)
	queue := []*object.Commit{c1}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if ancestors[c.Hash] {
			continue
		}
		ancestors[c.Hash] = true
		for _, p := range c.ParentHashes {
			if !ancestors[p] {
				pc, err := g.repo.CommitObject(p)
				if err == nil {
					queue = append(queue, pc)
				}
			}
		}
	}

	// Walk ancestors of h2 until we find one in h1's ancestry
	c2, err := g.repo.CommitObject(h2)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get commit2: %w", err)
	}

	queue2 := []*object.Commit{c2}
	visited := make(map[plumbing.Hash]bool)
	for len(queue2) > 0 {
		c := queue2[0]
		queue2 = queue2[1:]
		if visited[c.Hash] {
			continue
		}
		visited[c.Hash] = true

		if ancestors[c.Hash] {
			return c.Hash, nil
		}

		for _, p := range c.ParentHashes {
			if !visited[p] {
				pc, err := g.repo.CommitObject(p)
				if err == nil {
					queue2 = append(queue2, pc)
				}
			}
		}
	}

	return plumbing.ZeroHash, fmt.Errorf("no common ancestor found")
}
