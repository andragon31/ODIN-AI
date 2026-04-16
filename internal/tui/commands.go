// Package tui provides Völva - the interface engine for ODIN
package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Commands returns the Völva TUI CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "theme",
		Short: "Völva - Theme Engine",
		Long: `Völva is the interface engine for ODIN AI.
		
Provides:
- Multiple built-in themes (Rose Pine, Nord, Catppuccin Mocha, Dracula)
- Theme preview and switching
- Custom theme support`,
	}

	cmd.AddCommand(
		newThemeListCmd(),
		newThemeSetCmd(),
		newThemePreviewCmd(),
	)

	return cmd
}

// newThemeListCmd creates the 'theme list' command
func newThemeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available themes",
		Long:  `Lists all available themes with their names.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := NewThemeEngine()
			themes := engine.ListThemes()

			if jsonFlag, _ := cmd.Flags().GetBool("json"); jsonFlag {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(themes)
			}

			fmt.Println("╔══════════════════════════════════════════════════╗")
			fmt.Println("║           Völva - Available Themes              ║")
			fmt.Println("╠══════════════════════════════════════════════════╣")

			for i, theme := range themes {
				active := ""
				if theme.Name == engine.GetActiveTheme().Name {
					active = " (active)"
				}
				fmt.Printf("║  %d. %-30s%s%13s║\n", i+1, theme.Name, active, "║")
			}

			fmt.Println("╚══════════════════════════════════════════════════╝")

			return nil
		},
	}
}

// newThemeSetCmd creates the 'theme set' command
func newThemeSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <name>",
		Short: "Set the active theme",
		Long:  `Sets the active theme without requiring restart.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("theme name required")
			}

			themeName := args[0]
			engine := NewThemeEngine()

			if err := engine.SetActiveTheme(themeName); err != nil {
				return err
			}

			// Save theme preference to config
			homeDir, _ := os.UserHomeDir()
			themeConfig := filepath.Join(homeDir, ".odin", "theme.json")

			data := map[string]string{"current": themeName}
			jsonData, _ := json.MarshalIndent(data, "", "  ")

			os.MkdirAll(filepath.Dir(themeConfig), 0755)
			if err := os.WriteFile(themeConfig, jsonData, 0644); err != nil {
				return fmt.Errorf("failed to save theme: %w", err)
			}

			if jsonFlag, _ := cmd.Flags().GetBool("json"); jsonFlag {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]string{
					"status":  "success",
					"theme":   themeName,
					"message": "Theme updated - restart ODIN to apply to TUI",
				})
			}

			fmt.Printf("✓ Theme set to: %s\n", themeName)
			fmt.Println("  (Theme is applied immediately to output modes)")
			fmt.Println("  (Restart ODIN to apply to interactive TUI)")

			return nil
		},
	}
}

// newThemePreviewCmd creates the 'theme preview' command
func newThemePreviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preview [name]",
		Short: "Preview theme(s)",
		Long:  `Preview all themes or a specific theme with sample output.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := NewThemeEngine()

			if jsonFlag, _ := cmd.Flags().GetBool("json"); jsonFlag {
				themes := engine.ListThemes()
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(themes)
			}

			if len(args) > 0 {
				// Preview specific theme
				themeName := args[0]
				theme := engine.GetTheme(themeName)
				if theme == nil {
					return fmt.Errorf("theme not found: %s", themeName)
				}

				fmt.Println("╔══════════════════════════════════════════════════╗")
				fmt.Printf("║           Theme: %-28s       ║\n", themeName)
				fmt.Println("╚══════════════════════════════════════════════════╝")
				fmt.Println(RenderThemePreview(theme))
				return nil
			}

			// Preview all themes
			fmt.Println("╔══════════════════════════════════════════════════╗")
			fmt.Println("║           Völva - Theme Preview                 ║")
			fmt.Println("╚══════════════════════════════════════════════════╝")
			fmt.Println()

			themes := engine.ListThemes()
			for _, theme := range themes {
				fmt.Printf("═══ %s ═══\n", theme.Name)
				fmt.Println(RenderThemePreview(theme))
				fmt.Println()
			}

			return nil
		},
	}
}

// LoadSavedTheme loads the saved theme preference
func LoadSavedTheme() string {
	homeDir, _ := os.UserHomeDir()
	themeConfig := filepath.Join(homeDir, ".odin", "theme.json")

	data, err := os.ReadFile(themeConfig)
	if err != nil {
		return "rose-pine" // Default
	}

	var config map[string]string
	if err := json.Unmarshal(data, &config); err != nil {
		return "rose-pine"
	}

	if theme, ok := config["current"]; ok {
		return theme
	}

	return "rose-pine"
}
