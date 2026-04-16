package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/pkg/logger"
)

// Commands returns the plugin CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage ODIN plugins",
		Long: `Manage ODIN plugins - install, list, uninstall, and validate.
Plugins extend ODIN functionality through a secure WASM runtime.`,
	}

	cmd.AddCommand(
		newInstallCmd(),
		newListCmd(),
		newUninstallCmd(),
		newValidateCmd(),
	)

	return cmd
}

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <path>",
		Short: "Install a plugin from path",
		Long: `Install a plugin from a local path. The path should point to
a directory containing a manifest.json file.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(cmd, args[0])
		},
	}

	return cmd
}

func runInstall(cmd *cobra.Command, path string) error {
	runtime, err := NewRuntime(DefaultRuntimeConfig())
	if err != nil {
		return fmt.Errorf("failed to create runtime: %w", err)
	}

	// Resolve path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path is a manifest file or directory
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("plugin path not found: %w", err)
	}

	var manifestPath string
	if info.IsDir() {
		manifestPath = filepath.Join(absPath, "manifest.json")
	} else {
		manifestPath = absPath
	}

	if err := runtime.Install(manifestPath); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		data := map[string]string{
			"status": "success",
			"path":   absPath,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	fmt.Printf("Plugin installed successfully from: %s\n", absPath)
	return nil
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Long:  `Display all installed plugins with their versions and permissions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd)
		},
	}
}

func runList(cmd *cobra.Command) error {
	runtime, err := NewRuntime(DefaultRuntimeConfig())
	if err != nil {
		return fmt.Errorf("failed to create runtime: %w", err)
	}

	// List plugins from plugins directory
	pluginsDir := runtime.config.PluginsDir
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No plugins installed")
			return nil
		}
		return fmt.Errorf("failed to read plugins directory: %w", err)
	}

	plugins := []PluginMetadata{}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(pluginsDir, entry.Name(), "manifest.json")
		manifest, err := ValidatePluginManifest(manifestPath)
		if err != nil {
			logger.Warn("Invalid plugin manifest", "name", entry.Name(), "error", err)
			continue
		}

		plugins = append(plugins, PluginMetadata{
			Name:        manifest.Name,
			Version:     manifest.Version,
			Author:      manifest.Author,
			Description: manifest.Description,
			Permissions: manifest.Permissions,
		})
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(plugins)
	}

	if len(plugins) == 0 {
		fmt.Println("No plugins installed. Use 'odin plugin install <path>' to install one.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tAUTHOR\tPERMISSIONS")

	for _, p := range plugins {
		perms := ""
		if len(p.Permissions) > 0 {
			perms = p.Permissions[0]
			for i := 1; i < len(p.Permissions) && i < 3; i++ {
				perms += ", " + p.Permissions[i]
			}
			if len(p.Permissions) > 3 {
				perms += "..."
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, p.Version, p.Author, perms)
	}
	w.Flush()

	return nil
}

func newUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall <name>",
		Short: "Uninstall a plugin",
		Long: `Remove an installed plugin. This deletes the plugin files
from the plugins directory.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstall(cmd, args[0])
		},
	}

	return cmd
}

func runUninstall(cmd *cobra.Command, name string) error {
	runtime, err := NewRuntime(DefaultRuntimeConfig())
	if err != nil {
		return fmt.Errorf("failed to create runtime: %w", err)
	}

	if err := runtime.Uninstall(name); err != nil {
		return fmt.Errorf("failed to uninstall plugin: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		data := map[string]string{
			"status": "success",
			"name":   name,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	fmt.Printf("Plugin '%s' uninstalled successfully\n", name)
	return nil
}

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <path>",
		Short: "Validate a plugin manifest",
		Long: `Validate a plugin manifest without installing it.
Checks manifest syntax and required fields.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd, args[0])
		},
	}

	return cmd
}

func runValidate(cmd *cobra.Command, path string) error {
	// Resolve path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path is a manifest file or directory
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	var manifestPath string
	if info.IsDir() {
		manifestPath = filepath.Join(absPath, "manifest.json")
	} else {
		manifestPath = absPath
	}

	manifest, err := ValidatePluginManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{
			"valid":    true,
			"manifest": manifest,
		})
	}

	fmt.Printf("✓ Plugin manifest is valid\n")
	fmt.Printf("  Name:        %s\n", manifest.Name)
	fmt.Printf("  Version:     %s\n", manifest.Version)
	fmt.Printf("  Author:      %s\n", manifest.Author)
	fmt.Printf("  Entry Point: %s\n", manifest.EntryPoint)
	if len(manifest.Permissions) > 0 {
		fmt.Printf("  Permissions: %v\n", manifest.Permissions)
	}

	return nil
}
