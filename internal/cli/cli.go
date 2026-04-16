// Package cli provides the command-line interface for ODIN
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/internal/config"
	"github.com/odin-ai/odin/internal/deploy"
	"github.com/odin-ai/odin/internal/guardian"
	"github.com/odin-ai/odin/internal/memory"
	"github.com/odin-ai/odin/internal/migrate"
	"github.com/odin-ai/odin/internal/orchestrator"
	"github.com/odin-ai/odin/internal/plugins"
	"github.com/odin-ai/odin/internal/router"
	"github.com/odin-ai/odin/internal/skills"
	"github.com/odin-ai/odin/internal/sync"
	"github.com/odin-ai/odin/internal/tui"
	"github.com/odin-ai/odin/internal/verify"
	"github.com/odin-ai/odin/pkg/logger"
)

// NewRootCmd creates the root command
func NewRootCmd(version, buildTime string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "odin",
		Short: "ODIN AI - Nórdico Local-First AI Ecosystem",
		Long: `ODIN AI es el orquestador local-first del ecosistema nórdico.

Completamente offline, 100% OSS, sin costos de infraestructura.
Inspirado en Gentleman AI, potenciado para local-first.`,
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if quiet, _ := cmd.Flags().GetBool("quiet"); quiet {
				logger.SetLevel(logger.ErrorLevel)
			} else if debug, _ := cmd.Flags().GetBool("debug"); debug {
				logger.SetLevel(logger.DebugLevel)
			}
		},
	}

	cmd.PersistentFlags().Bool("quiet", false, "Suppress output (for CI/scripts)")
	cmd.PersistentFlags().Bool("debug", false, "Enable debug output")
	cmd.PersistentFlags().Bool("json", false, "Output in JSON format")
	cmd.PersistentFlags().String("config", "", "Custom config path")

	cmd.AddCommand(
		newInitCmd(),
		newStatusCmd(version),
		newVersionCmd(version, buildTime),
		newConfigCmd(),
		newSessionCmd(),
		memory.Commands(),
		router.Commands(),
		guardian.Commands(),
		sync.SyncCommands(),
		skills.Commands(),
		deploy.Commands(),
		verify.Commands(),
		tui.Commands(),
		migrate.Commands(),
		plugins.Commands(),
		orchestrator.Commands(),
	)

	return cmd
}

// initCmd represents the init command
func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize ODIN environment",
		Long:  `Initialize the ODIN configuration and directory structure.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return runInit(ctx, cmd)
		},
	}
}

func runInit(ctx context.Context, cmd *cobra.Command) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		data := map[string]string{
			"status":   "success",
			"message":  "ODIN initialized",
			"home_dir": cfg.HomeDir,
			"mode":     cfg.Mode,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	logger.Info("ODIN initialized successfully",
		"version", "1.0",
		"home_dir", cfg.HomeDir,
		"mode", cfg.Mode,
	)

	return nil
}

// statusCmd represents the status command
func newStatusCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show ODIN ecosystem status",
		Long:  `Display the health status of all ODIN components.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd)
		},
	}
}

func runStatus(cmd *cobra.Command) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Check if Mimir is initialized
	cfg := memory.DefaultConfig()
	mimirState := "not_initialized"
	if _, err := os.Stat(cfg.DBPath); err == nil {
		mimirState = "healthy"
	}

	// Check if Guardian (Heimdall) is initialized
	guardianState := "not_initialized"
	guardianCfg := config.DefaultConfig()
	if _, err := os.Stat(guardianCfg.Guardian.RulesPath); err == nil {
		guardianState = "healthy"
	}

	// Check if Bifrost is initialized
	bifrostState := "not_initialized"
	bifrostRepoPath := sync.DefaultRepoPath()
	if _, err := os.Stat(filepath.Join(bifrostRepoPath, ".git")); err == nil {
		bifrostState = "healthy"
	}

	// Check if Runes is initialized
	runesState := "not_initialized"
	runesPath := skills.DefaultRunesPath()
	if _, err := os.Stat(runesPath); err == nil {
		runesState = "healthy"
	}

	// Check if Dvergar is initialized
	dvergarState := "not_initialized"
	dvergarCfg := deploy.DefaultDeployConfig()
	dvergar := deploy.New(dvergarCfg)
	if dvergar.IsInstalled() {
		dvergarState = "healthy"
	}

	// Check if Nornir is initialized
	nornirState := "not_initialized"
	odinCfg := config.DefaultConfig()
	if _, err := os.Stat(filepath.Join(odinCfg.HomeDir, ".odin", "verify")); err == nil {
		nornirState = "healthy"
	}

	// Check if Völva is initialized
	volvaState := "not_initialized"
	if _, err := os.Stat(filepath.Join(odinCfg.HomeDir, ".odin", "theme.json")); err == nil {
		volvaState = "healthy"
	}

	// Check if Migrate is initialized
	migrateState := "not_initialized"
	if _, err := os.Stat(filepath.Join(odinCfg.HomeDir, ".odin", "config")); err == nil {
		migrateState = "healthy"
	}

	// Check if Plugins is initialized
	pluginsState := "not_initialized"
	pluginsPath := filepath.Join(odinCfg.HomeDir, ".odin", "plugins")
	if _, err := os.Stat(pluginsPath); err == nil {
		pluginsState = "healthy"
	}

	// Check if Orchestrator is initialized
	orchestratorState := "not_initialized"
	sessionsPath := filepath.Join(odinCfg.HomeDir, ".odin", "sessions")
	if _, err := os.Stat(sessionsPath); err == nil {
		orchestratorState = "healthy"
	}

	status := map[string]interface{}{
		"version": "1.0.0",
		"status":  "healthy",
		"components": map[string]interface{}{
			"odin":         "healthy",
			"mimir":        mimirState,
			"heimdall":     guardianState,
			"bifrost":      bifrostState,
			"runes":        runesState,
			"nornir":       nornirState,
			"dvergar":      dvergarState,
			"volva":        volvaState,
			"migrate":      migrateState,
			"plugins":      pluginsState,
			"orchestrator": orchestratorState,
		},
		"mode": "local",
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(status)
	}

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║           ODIN AI - Status Report              ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf("║  Version: %-37s║\n", status["version"])
	fmt.Printf("║  Status:  %-37s║\n", status["status"])
	fmt.Printf("║  Mode:     %-37s║\n", status["mode"])
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Println("║  Components:                                     ║")

	components := status["components"].(map[string]interface{})
	for name, state := range components {
		fmt.Printf("║    %-10s %-27s║\n", name+":", state)
	}

	fmt.Println("╚══════════════════════════════════════════════════╝")

	return nil
}

// versionCmd represents the version command
func newVersionCmd(version, buildTime string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show ODIN version",
		Long:  `Display version and build information.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ODIN AI v%s\n", version)
			fmt.Printf("Build time: %s\n", buildTime)
		},
	}
}

// configCmd represents the config command
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage ODIN configuration",
		Long:  `View or modify ODIN configuration.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("")
			if err != nil {
				return err
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(cfg)
		},
	})

	return cmd
}

// sessionCmd represents session management commands
func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage ODIN sessions",
		Long:  `List, resume, or manage SDD sessions.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List active sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("No active sessions")
			return nil
		},
	})

	return cmd
}

// TUIView represents the main TUI view
type TUIView struct {
	quitting bool
}

func NewTUIView() *TUIView {
	return &TUIView{}
}

func (t *TUIView) Init() tea.Cmd {
	return nil
}

func (t *TUIView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			t.quitting = true
			return t, tea.Quit
		}
	}
	return t, nil
}

func (t *TUIView) View() string {
	if t.quitting {
		return "Goodbye!\n"
	}
	return "ODIN AI - Nórdico Local-First\nPress 'q' to quit.\n"
}
