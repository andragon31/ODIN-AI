// Package catalog provides the component catalog system for ODIN
package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/pkg/logger"
)

// Commands returns all catalog CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Catalog - List and install ODIN components",
		Long: `Catalog system for ODIN AI ecosystem.
Lists known AI agents, installable components, and available runes.
Use to discover and install ODIN ecosystem components.`,
	}

	cmd.AddCommand(
		newListCmd(),
		newInstallCmd(),
		newInfoCmd(),
	)

	return cmd
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List catalog items",
		Long: `List all agents, components, or runes in the catalog.
Use --type to filter by category.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd)
		},
	}

	cmd.Flags().String("type", "", "Filter by type: agents, components, runes")

	return cmd
}

func runList(cmd *cobra.Command) error {
	itemType, _ := cmd.Flags().GetString("type")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	manager := DefaultCatalogManager()

	if itemType == "" || itemType == "agents" {
		if itemType == "" {
			fmt.Println("\n=== AI Agents ===")
		}
		agents := manager.ListAgents()
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(agents)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION")
		for _, a := range agents {
			fmt.Fprintf(w, "%s\t%s\t%s\n", a.ID, a.Name, a.Description)
		}
		w.Flush()
	}

	if itemType == "" || itemType == "components" {
		if itemType == "" {
			fmt.Println("\n=== Components ===")
		}
		components := manager.ListComponents()
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(components)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tVERSION\tDESCRIPTION")
		for _, c := range components {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.ID, c.Name, c.Version, c.Description)
		}
		w.Flush()
	}

	if itemType == "" || itemType == "runes" {
		if itemType == "" {
			fmt.Println("\n=== Runes ===")
		}
		runes := manager.ListRunes()
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(runes)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDESCRIPTION\tTAGS")
		for _, r := range runes {
			fmt.Fprintf(w, "%s\t%s\t%s\n", r.Name, r.Description, joinTags(r.Tags))
		}
		w.Flush()
	}

	return nil
}

func joinTags(tags []string) string {
	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += ", "
		}
		result += tag
	}
	return result
}

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <component>",
		Short: "Install a component from catalog",
		Long: `Install a component using the ODIN pipeline with automatic
backup and rollback support. Example: odin catalog install sdd`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("component name required")
			}
			return runInstall(cmd, args[0])
		},
	}

	return cmd
}

func runInstall(cmd *cobra.Command, componentID string) error {
	manager := DefaultCatalogManager()

	comp := manager.GetComponent(componentID)
	if comp == nil {
		return fmt.Errorf("component %s not found in catalog", componentID)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")

	if jsonOutput {
		data := map[string]interface{}{
			"status":      "installing",
			"component":   comp.ID,
			"name":        comp.Name,
			"description": comp.Description,
			"depends_on":  comp.DependsOn,
			"runes":       comp.Runes,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	logger.Info("Installing component via pipeline",
		"component", comp.ID,
		"name", comp.Name,
		"runes", comp.Runes,
	)

	fmt.Printf("Installing %s (%s)...\n", comp.Name, comp.ID)
	fmt.Println("Use 'odin pipeline install " + comp.ID + "' for full pipeline with backup and rollback")

	return nil
}

func newInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <name>",
		Short: "Show detailed information about a catalog item",
		Long: `Show detailed information about an agent, component, or rune.
Example: odin catalog info sdd`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("name required")
			}
			return runInfo(cmd, args[0])
		},
	}

	return cmd
}

func runInfo(cmd *cobra.Command, name string) error {
	manager := DefaultCatalogManager()
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Check agents first
	if agent := manager.GetAgent(AgentID(name)); agent != nil {
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(agent)
		}
		fmt.Println("\n=== Agent Information ===")
		fmt.Printf("ID:          %s\n", agent.ID)
		fmt.Printf("Name:        %s\n", agent.Name)
		fmt.Printf("Description: %s\n", agent.Description)
		fmt.Printf("Website:     %s\n", agent.Website)
		fmt.Printf("Supported:   %s\n", joinSlice(agent.SupportedOS))
		return nil
	}

	// Check components
	if comp := manager.GetComponent(name); comp != nil {
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(comp)
		}
		fmt.Println("\n=== Component Information ===")
		fmt.Printf("ID:          %s\n", comp.ID)
		fmt.Printf("Name:        %s\n", comp.Name)
		fmt.Printf("Version:     %s\n", comp.Version)
		fmt.Printf("Description: %s\n", comp.Description)
		fmt.Printf("Depends On:  %s\n", joinSlice(comp.DependsOn))
		fmt.Printf("Runes:       %s\n", joinSlice(comp.Runes))
		return nil
	}

	// Check runes
	if rune := manager.GetRune(name); rune != nil {
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(rune)
		}
		fmt.Println("\n=== Rune Information ===")
		fmt.Printf("Name:        %s\n", rune.Name)
		fmt.Printf("Description: %s\n", rune.Description)
		fmt.Printf("Tags:        %s\n", joinTags(rune.Tags))
		return nil
	}

	return fmt.Errorf("'%s' not found in catalog", name)
}

func joinSlice(slice []string) string {
	if len(slice) == 0 {
		return "none"
	}
	result := ""
	for i, s := range slice {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
