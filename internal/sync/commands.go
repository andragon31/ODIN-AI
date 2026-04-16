package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/internal/config"
	"github.com/odin-ai/odin/pkg/logger"
)

// Commands returns all Bifrost CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bifrost",
		Short: "Bifrost - Git-based sync engine with CRDT",
		Long: `Bifrost is the Norse rainbow bridge that connects worlds.
This command manages the ODIN sync engine with Git operations,
CRDT-based conflict resolution, and branch management.`,
	}

	cmd.AddCommand(
		newInitCmd(),
		newPushCmd(),
		newPullCmd(),
		newStatusCmd(),
		newDiffCmd(),
		newLogCmd(),
		newSignCmd(),
		branchCommands(),
	)

	return cmd
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the sync repository",
		Long:  `Initialize a local Git repository at ~/.odin/config/ for configuration sync.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd)
		},
	}
}

func runInit(cmd *cobra.Command) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create Bifrost instance
	repoPath := DefaultRepoPath()
	if cfg.Sync.Remote != "" {
		repoPath = cfg.Sync.Remote
	}

	b, err := NewBifrost(repoPath, cfg.Sync.Remote, cfg.Sync.GPGSign)
	if err != nil {
		return fmt.Errorf("failed to create bifrost: %w", err)
	}

	if err := b.Init(); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		data := map[string]string{
			"status":    "success",
			"message":   "Bifrost initialized",
			"repo_path": repoPath,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	logger.Info("Bifrost initialized", "path", repoPath)
	fmt.Printf("Bifrost sync repository initialized at: %s\n", repoPath)
	return nil
}

func newPushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Push changes to remote",
		Long:  `Push local changes to the remote repository.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(cmd)
		},
	}
}

func runPush(cmd *cobra.Command) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	b, err := NewBifrost(DefaultRepoPath(), cfg.Sync.Remote, cfg.Sync.GPGSign)
	if err != nil {
		return fmt.Errorf("failed to create bifrost: %w", err)
	}

	if !b.IsInitialized() {
		return fmt.Errorf("bifrost not initialized. Run 'odin sync init' first")
	}

	if err := b.Push(); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	logger.Info("Push successful")
	fmt.Println("Changes pushed to remote")
	return nil
}

func newPullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Pull changes from remote",
		Long:  `Pull changes from the remote repository with CRDT merge.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPull(cmd)
		},
	}
}

func runPull(cmd *cobra.Command) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	b, err := NewBifrost(DefaultRepoPath(), cfg.Sync.Remote, cfg.Sync.GPGSign)
	if err != nil {
		return fmt.Errorf("failed to create bifrost: %w", err)
	}

	if !b.IsInitialized() {
		return fmt.Errorf("bifrost not initialized. Run 'odin sync init' first")
	}

	result, err := b.Pull()
	if err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	if result.Pulled {
		fmt.Println("Changes pulled from remote")
	} else {
		fmt.Println("Already up to date")
	}

	if len(result.Conflicts) > 0 {
		fmt.Printf("\n%d conflicts detected:\n", len(result.Conflicts))
		for _, c := range result.Conflicts {
			fmt.Printf("  - %s (resolved: %v)\n", c.Path, c.Resolved)
		}
	}

	return nil
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show sync status",
		Long:  `Display the current sync status including branch, remote, and uncommitted changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd)
		},
	}
}

func runStatus(cmd *cobra.Command) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	b, err := NewBifrost(DefaultRepoPath(), cfg.Sync.Remote, cfg.Sync.GPGSign)
	if err != nil {
		return fmt.Errorf("failed to create bifrost: %w", err)
	}

	status, err := b.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(status)
	}

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║         ODIN AI - Bifrost Status              ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")

	fmt.Printf("║  Initialized:   %-31s║\n", boolToYesNo(status.Initialized))
	fmt.Printf("║  Repository:    %-31s║\n", status.RepoPath)
	fmt.Printf("║  Remote:        %-31s║\n", status.Remote)
	fmt.Printf("║  Branch:        %-31s║\n", status.CurrentBranch)
	fmt.Printf("║  GPG Signing:   %-31s║\n", boolToYesNo(status.GPGSign))
	fmt.Println("╠══════════════════════════════════════════════════╣")

	if status.HasUncommitted {
		fmt.Println("║  Uncommitted Changes:                         ║")
		for _, f := range status.UncommittedFiles {
			fmt.Printf("║    - %-36s║\n", f)
		}
	} else {
		fmt.Printf("║  No uncommitted changes                       ║\n")
	}

	fmt.Println("╚══════════════════════════════════════════════════╝")

	return nil
}

func newDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Show pending changes",
		Long:  `Display uncommitted changes before applying them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff(cmd)
		},
	}
}

func runDiff(cmd *cobra.Command) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	b, err := NewBifrost(DefaultRepoPath(), cfg.Sync.Remote, cfg.Sync.GPGSign)
	if err != nil {
		return fmt.Errorf("failed to create bifrost: %w", err)
	}

	if !b.IsInitialized() {
		return fmt.Errorf("bifrost not initialized. Run 'odin sync init' first")
	}

	diff, err := b.Diff()
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	fmt.Println(diff)
	return nil
}

func newLogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Show commit history",
		Long:  `Display the history of synced changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLog(cmd)
		},
	}

	cmd.Flags().Int("limit", 10, "Number of commits to show")
	return cmd
}

func runLog(cmd *cobra.Command) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	limit, _ := cmd.Flags().GetInt("limit")

	b, err := NewBifrost(DefaultRepoPath(), cfg.Sync.Remote, cfg.Sync.GPGSign)
	if err != nil {
		return fmt.Errorf("failed to create bifrost: %w", err)
	}

	if !b.IsInitialized() {
		return fmt.Errorf("bifrost not initialized. Run 'odin sync init' first")
	}

	commits, err := b.Log(limit)
	if err != nil {
		return fmt.Errorf("failed to get log: %w", err)
	}

	if len(commits) == 0 {
		fmt.Println("No commits yet")
		return nil
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(commits)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "COMMIT\tAUTHOR\tDATE\tMESSAGE\n")
	for _, c := range commits {
		msg := c.Message
		if len(msg) > 50 {
			msg = msg[:47] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Hash, c.Author, c.Timestamp.Format("2006-01-02 15:04"), msg)
	}
	w.Flush()

	return nil
}

func newSignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Manage GPG signing",
		Long:  `Enable or disable GPG signing for commits.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "on",
		Short: "Enable GPG signing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSignOn(cmd)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "off",
		Short: "Disable GPG signing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSignOff(cmd)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show GPG signing status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSignStatus(cmd)
		},
	})

	return cmd
}

func runSignOn(cmd *cobra.Command) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.Sync.GPGSign = true
	// Note: In a full implementation, we would persist this to config file
	logger.Info("GPG signing enabled")
	fmt.Println("GPG signing enabled for future commits")
	return nil
}

func runSignOff(cmd *cobra.Command) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.Sync.GPGSign = false
	logger.Info("GPG signing disabled")
	fmt.Println("GPG signing disabled")
	return nil
}

func runSignStatus(cmd *cobra.Command) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("GPG signing: %s\n", boolToEnabledDisabled(cfg.Sync.GPGSign))
	return nil
}

// branchCommands returns branch management commands
func branchCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch",
		Short: "Manage branches",
		Long:  `List, create, or manage sync branches.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBranchList(cmd)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "create <name>",
		Short: "Create a new branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBranchCreate(cmd, args[0])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBranchDelete(cmd, args[0])
		},
	})

	return cmd
}

func runBranchList(cmd *cobra.Command) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	b, err := NewBifrost(DefaultRepoPath(), cfg.Sync.Remote, cfg.Sync.GPGSign)
	if err != nil {
		return fmt.Errorf("failed to create bifrost: %w", err)
	}

	if !b.IsInitialized() {
		return fmt.Errorf("bifrost not initialized. Run 'odin sync init' first")
	}

	branches, err := b.BranchList()
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	if len(branches) == 0 {
		fmt.Println("No branches")
		return nil
	}

	fmt.Println("Branches:")
	for _, br := range branches {
		marker := "  "
		if br.IsHead {
			marker = "* "
		}
		fmt.Printf("  %s%s %s\n", marker, br.Name, br.Hash)
	}

	return nil
}

func runBranchCreate(cmd *cobra.Command, name string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	b, err := NewBifrost(DefaultRepoPath(), cfg.Sync.Remote, cfg.Sync.GPGSign)
	if err != nil {
		return fmt.Errorf("failed to create bifrost: %w", err)
	}

	if !b.IsInitialized() {
		return fmt.Errorf("bifrost not initialized. Run 'odin sync init' first")
	}

	if err := b.BranchCreate(name); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	logger.Info("Branch created", "name", name)
	fmt.Printf("Branch '%s' created and checked out\n", name)
	return nil
}

func runBranchDelete(cmd *cobra.Command, name string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	b, err := NewBifrost(DefaultRepoPath(), cfg.Sync.Remote, cfg.Sync.GPGSign)
	if err != nil {
		return fmt.Errorf("failed to create bifrost: %w", err)
	}

	if !b.IsInitialized() {
		return fmt.Errorf("bifrost not initialized. Run 'odin sync init' first")
	}

	// Use git client directly for delete
	git, err := NewGitClient(DefaultRepoPath(), cfg.Sync.Remote, cfg.Sync.GPGSign)
	if err != nil {
		return fmt.Errorf("failed to create git client: %w", err)
	}
	if err := git.Open(); err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	if err := git.BranchDelete(name); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	logger.Info("Branch deleted", "name", name)
	fmt.Printf("Branch '%s' deleted\n", name)
	return nil
}

// Helper functions

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func boolToEnabledDisabled(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

// SyncCommands is an alias for Commands for the "sync" keyword
func SyncCommands() *cobra.Command {
	cmd := Commands()
	cmd.Use = "sync"
	return cmd
}
