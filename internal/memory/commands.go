// Package memory provides the Mimir memory engine for ODIN
package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/pkg/logger"
)

// Commands returns all Mimir CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mimir",
		Short: "Mimir - Memory engine with semantic search",
		Long: `Mimir is the Norse god of wisdom and knowledge.
This command manages the ODIN memory engine with semantic search,
encryption, and knowledge graph capabilities.`,
	}

	cmd.AddCommand(
		newStoreCmd(),
		newSearchCmd(),
		newRecallCmd(),
		newTagsCmd(),
		newPruneCmd(),
		newEncryptCmd(),
		newDecryptCmd(),
		newSyncCmd(),
		newGraphCmd(),
		newListCmd(),
		newDeleteCmd(),
	)

	return cmd
}

func newStoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store <content>",
		Short: "Store a new memory",
		Long:  `Store a new memory with automatic embedding generation.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content := args[0]
			project, _ := cmd.Flags().GetString("project")
			tagsStr, _ := cmd.Flags().GetString("tags")

			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			// Parse tags
			var tags []string
			if tagsStr != "" {
				tags = parseTags(tagsStr)
			}

			m := &Memory{
				Content: content,
				Project: project,
				Tags:    tags,
			}

			if err := store.Store(m); err != nil {
				return fmt.Errorf("failed to store memory: %w", err)
			}

			jsonOutput, _ := cmd.Flags().GetBool("json")
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(m)
			}

			logger.Info("Memory stored", "id", m.ID)
			fmt.Printf("Memory stored with ID: %s\n", m.ID)
			return nil
		},
	}
	cmd.Flags().String("project", "", "Project filter")
	cmd.Flags().String("tags", "", "Comma-separated tags")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	return cmd
}

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search --query <text>",
		Short: "Search memories semantically",
		Long:  `Search memories using semantic vector search or FTS5 fallback.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			query, _ := cmd.Flags().GetString("query")
			limit, _ := cmd.Flags().GetInt("limit")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			if query == "" && len(args) > 0 {
				query = args[0]
			}
			if query == "" {
				return fmt.Errorf("query is required")
			}

			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			results, err := store.Search(query, limit)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(results)
			}

			// Human-readable output
			if len(results) == 0 {
				fmt.Println("No results found")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "ID\tSCORE\tTAGS\tCONTENT\n")
			for _, r := range results {
				tags := ""
				if len(r.Memory.Tags) > 0 {
					tags = joinTags(r.Memory.Tags)
				}
				content := r.Memory.Content
				if len(content) > 50 {
					content = content[:47] + "..."
				}
				fmt.Fprintf(w, "%s\t%.2f\t%s\t%s\n", r.Memory.ID, r.Score, tags, content)
			}
			w.Flush()

			return nil
		},
	}
	cmd.Flags().String("query", "", "Search query")
	cmd.Flags().Int("limit", 10, "Maximum results to return")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	return cmd
}

func newRecallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recall --id <memory-id>",
		Short: "Recall a memory by ID",
		Long:  `Retrieve a specific memory by its ID.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, _ := cmd.Flags().GetString("id")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			if id == "" {
				if len(args) < 1 {
					return fmt.Errorf("memory ID is required")
				}
				id = args[0]
			}

			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			m, err := store.Recall(id)
			if err != nil {
				return fmt.Errorf("failed to recall memory: %w", err)
			}
			if m == nil {
				return fmt.Errorf("memory not found: %s", id)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(m)
			}

			// Human-readable output
			fmt.Printf("ID:       %s\n", m.ID)
			fmt.Printf("Project:  %s\n", m.Project)
			fmt.Printf("Tags:     %s\n", joinTags(m.Tags))
			fmt.Printf("Created:  %s\n", m.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Updated:  %s\n", m.UpdatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Accessed: %s\n", m.AccessedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("\nContent:\n%s\n", m.Content)

			return nil
		},
	}
	cmd.Flags().String("id", "", "Memory ID to recall")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	return cmd
}

func newTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Manage memory tags",
		Long:  `List available tags or add tags to a memory.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all available tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			tags, err := store.ListTags()
			if err != nil {
				return fmt.Errorf("failed to list tags: %w", err)
			}

			if len(tags) == 0 {
				fmt.Println("No tags found")
				return nil
			}

			fmt.Println("Available tags:")
			for _, tag := range tags {
				fmt.Printf("  - %s\n", tag)
			}

			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "add <memory-id> <tag>",
		Short: "Add a tag to a memory",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			memoryID := args[0]
			tag := args[1]

			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			if err := store.AddTag(memoryID, tag); err != nil {
				return fmt.Errorf("failed to add tag: %w", err)
			}

			logger.Info("Tag added", "memory_id", memoryID, "tag", tag)
			fmt.Printf("Tag '%s' added to memory %s\n", tag, memoryID)
			return nil
		},
	})

	return cmd
}

func newPruneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune [--keep-tags arch,spec,security]",
		Short: "Prune memories not matching keep tags",
		Long: `Remove memories that don't have any of the specified tags.
By default, keeps memories with tags: arch, spec, security.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			keepTagsStr, _ := cmd.Flags().GetString("keep-tags")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			var keepTags []string
			if keepTagsStr != "" {
				keepTags = parseTags(keepTagsStr)
			}

			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			if dryRun {
				// Dry run mode
				smartPruner := NewSmartPruner(store.projectDB, keepTags, 0)
				report, err := smartPruner.PruningDryRun()
				if err != nil {
					return fmt.Errorf("dry run failed: %w", err)
				}

				if jsonOutput {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(report)
				}

				fmt.Printf("Would prune %d memories, keeping %d\n",
					report.TotalPruned, len(report.ProtectedMemories))
				return nil
			}

			count, err := store.Prune(keepTags)
			if err != nil {
				return fmt.Errorf("prune failed: %w", err)
			}

			if jsonOutput {
				data := map[string]int{"deleted": count}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(data)
			}

			logger.Info("Pruning completed", "deleted", count)
			fmt.Printf("Pruned %d memories\n", count)
			return nil
		},
	}
	cmd.Flags().String("keep-tags", "arch,spec,security", "Comma-separated tags to keep")
	cmd.Flags().Bool("dry-run", false, "Show what would be pruned without deleting")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	return cmd
}

func newEncryptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypt --key <key-file>",
		Short: "Encrypt the memory database",
		Long:  `Encrypt the memory database using age-compatible encryption.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyFile, _ := cmd.Flags().GetString("key")

			if keyFile == "" {
				return fmt.Errorf("key file is required")
			}

			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			if err := store.Encrypt(keyFile); err != nil {
				return fmt.Errorf("encryption failed: %w", err)
			}

			logger.Info("Database encrypted", "key_file", keyFile)
			fmt.Printf("Database encrypted successfully\n")
			return nil
		},
	}
	cmd.Flags().String("key", "", "Key file for encryption")
	return cmd
}

func newDecryptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decrypt --key <key-file>",
		Short: "Decrypt the memory database",
		Long:  `Decrypt the memory database using age-compatible encryption.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyFile, _ := cmd.Flags().GetString("key")

			if keyFile == "" {
				return fmt.Errorf("key file is required")
			}

			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			if err := store.Decrypt(keyFile); err != nil {
				return fmt.Errorf("decryption failed: %w", err)
			}

			logger.Info("Database decrypted", "key_file", keyFile)
			fmt.Printf("Database decrypted successfully\n")
			return nil
		},
	}
	cmd.Flags().String("key", "", "Key file for decryption")
	return cmd
}

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize with remote",
		Long:  `Push or pull memories from a remote storage.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "push",
		Short: "Push to remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			remote, _ := cmd.Flags().GetString("remote")

			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			if err := store.SyncPush(remote); err != nil {
				return fmt.Errorf("push failed: %w", err)
			}

			logger.Info("Push completed", "remote", remote)
			fmt.Printf("Pushed to remote: %s\n", remote)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "pull",
		Short: "Pull from remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			remote, _ := cmd.Flags().GetString("remote")

			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			if err := store.SyncPull(remote); err != nil {
				return fmt.Errorf("pull failed: %w", err)
			}

			logger.Info("Pull completed", "remote", remote)
			fmt.Printf("Pulled from remote: %s\n", remote)
			return nil
		},
	})

	return cmd
}

func newGraphCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph --from <memory-id>",
		Short: "Query the knowledge graph",
		Long:  `Query the knowledge graph starting from a memory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fromID, _ := cmd.Flags().GetString("from")
			depth, _ := cmd.Flags().GetInt("depth")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			if fromID == "" {
				if len(args) < 1 {
					return fmt.Errorf("memory ID is required")
				}
				fromID = args[0]
			}

			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			memories, err := store.Graph(fromID, depth)
			if err != nil {
				return fmt.Errorf("graph query failed: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(memories)
			}

			// Human-readable output
			fmt.Printf("Knowledge graph from memory %s (depth=%d):\n\n", fromID, depth)
			for _, m := range memories {
				tags := ""
				if len(m.Tags) > 0 {
					tags = fmt.Sprintf("[%s]", joinTags(m.Tags))
				}
				content := m.Content
				if len(content) > 60 {
					content = content[:57] + "..."
				}
				fmt.Printf("  %s %s\n    %s\n\n", m.ID, tags, content)
			}

			return nil
		},
	}
	cmd.Flags().String("from", "", "Starting memory ID")
	cmd.Flags().Int("depth", 3, "Traversal depth")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	return cmd
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [--project <name>]",
		Short: "List all memories",
		Long:  `List all memories, optionally filtered by project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			project, _ := cmd.Flags().GetString("project")
			limit, _ := cmd.Flags().GetInt("limit")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			memories, err := store.ListMemories(project)
			if err != nil {
				return fmt.Errorf("failed to list memories: %w", err)
			}

			if limit > 0 && len(memories) > limit {
				memories = memories[:limit]
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(memories)
			}

			// Human-readable output
			if len(memories) == 0 {
				fmt.Println("No memories found")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "ID\tPROJECT\tTAGS\tCREATED\n")
			for _, m := range memories {
				tags := ""
				if len(m.Tags) > 0 {
					tags = joinTags(m.Tags)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					m.ID, m.Project, tags, m.CreatedAt.Format("2006-01-02"))
			}
			w.Flush()

			fmt.Printf("\nTotal: %d memories\n", len(memories))
			return nil
		},
	}
	cmd.Flags().String("project", "", "Project filter")
	cmd.Flags().Int("limit", 0, "Limit results")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	return cmd
}

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <memory-id>",
		Short: "Delete a memory",
		Long:  `Delete a memory by its ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			cfg := DefaultConfig()
			store, err := NewStore(cfg)
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}
			defer store.Close()

			if err := store.Delete(id); err != nil {
				return fmt.Errorf("failed to delete memory: %w", err)
			}

			logger.Info("Memory deleted", "id", id)
			fmt.Printf("Memory %s deleted\n", id)
			return nil
		},
	}
}

// Helper functions

func parseTags(tagsStr string) []string {
	var tags []string
	parts := splitAndTrim(tagsStr, ",")
	for _, p := range parts {
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}

func splitAndTrim(s, sep string) []string {
	var result []string
	parts := strings.Split(s, sep)
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

func joinTags(tags []string) string {
	result := ""
	for i, t := range tags {
		if i > 0 {
			result += ", "
		}
		result += t
	}
	return result
}

// AddFlags adds common flags to commands
func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().String("project", "", "Project filter")
}

// init adds common flags to subcommands
func init() {
	// This function is intentionally empty - flags are added in individual commands
}

var _ = splitAndTrim // suppress unused warning
var _ = joinTags     // suppress unused warning
