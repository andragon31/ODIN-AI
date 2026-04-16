package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/pkg/logger"
)

// Commands returns all agents CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install AI agent configurations",
		Long: `Install and manage configurations for AI coding agents.
Supports Claude Code, Gemini CLI, Cursor, Windsurf, OpenCode, and Codex.
If the agent CLI is not installed, ODIN uses its local Router (Ollama)
to generate config content without external dependencies.`,
	}

	cmd.AddCommand(
		newDetectCmd(),
		newInstallAgentCmd(),
		newListAgentsCmd(),
		newVerifyAgentCmd(),
		newUninstallAgentCmd(),
		newInstallAllCmd(),
	)

	return cmd
}

// newDetectCmd creates the detect command
func newDetectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detect",
		Short: "Detect installed AI agents",
		Long:  `Detect which AI agents are installed on the system.`,
		RunE:  runDetect,
	}

	cmd.Flags().Bool("fast", false, "Skip version detection for faster results")

	return cmd
}

func runDetect(cmd *cobra.Command, args []string) error {
	fast, _ := cmd.Flags().GetBool("fast")
	detector := NewDetector()

	var results []DetectionResult
	if fast {
		results = detector.DetectAgentsFast()
	} else {
		results = detector.DetectAll()
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	PrintDetectionReport(results)
	return nil
}

// newInstallAgentCmd creates the install command for a specific agent
func newInstallAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent [agent-id]",
		Short: "Install configuration for a specific agent",
		Long: `Install configuration for a specific AI agent.
Available agents: claude-code, gemini-cli, cursor, windsurf, opencode, codex`,
		Args: cobra.ExactArgs(1),
		RunE: runInstallAgent,
	}

	cmd.Flags().String("model", "", "Model to use (e.g., claude-3-5-sonnet)")
	cmd.Flags().String("rules-path", "", "Custom rules path")
	cmd.Flags().String("config-path", "", "Custom config path")

	return cmd
}

func runInstallAgent(cmd *cobra.Command, args []string) error {
	agentID := args[0]

	agent := DetectCLIByName(agentID)
	if agent == nil {
		return fmt.Errorf("unknown agent: %s. Use 'odin install detect' to see available agents", agentID)
	}

	cfg := &AgentConfig{
		Model: getStringFlag(cmd, "model"),
	}

	if agent.Available() {
		logger.Info("Agent CLI is installed, using local generation", "agent", agent.Name())
	} else {
		logger.Info("Agent CLI not installed, using ODIN Router for config generation", "agent", agent.Name())
	}

	if err := agent.Install(cfg); err != nil {
		return fmt.Errorf("failed to install %s: %w", agent.Name(), err)
	}

	fmt.Printf("✓ %s configuration installed successfully\n", agent.Name())

	// Verify installation
	if err := agent.Verify(); err != nil {
		logger.Warn("Installation verification failed", "error", err)
	}

	return nil
}

// newListAgentsCmd creates the list command
func newListAgentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all supported agents",
		Long:  `List all supported AI agents with their status.`,
		RunE:  runListAgents,
	}
}

func runListAgents(cmd *cobra.Command, args []string) error {
	detector := NewDetector()
	results := detector.DetectAgentsFast()

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tAVAILABLE\tCONFIGURED")

	for _, r := range results {
		available := "✗"
		if r.Available {
			available = "✓"
		}
		configured := "✗"
		if r.ConfigExists {
			configured = "✓"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.AgentID, r.Name, available, configured)
	}
	w.Flush()

	return nil
}

// newVerifyAgentCmd creates the verify command
func newVerifyAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify [agent-id]",
		Short: "Verify agent configuration",
		Long:  `Verify that an agent's configuration is properly installed.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runVerifyAgent,
	}

	return cmd
}

func runVerifyAgent(cmd *cobra.Command, args []string) error {
	agentID := args[0]

	agent := DetectCLIByName(agentID)
	if agent == nil {
		return fmt.Errorf("unknown agent: %s", agentID)
	}

	if err := agent.Verify(); err != nil {
		return fmt.Errorf("verification failed for %s: %w", agent.Name(), err)
	}

	fmt.Printf("✓ %s configuration is valid\n", agent.Name())
	return nil
}

// newUninstallAgentCmd creates the uninstall command
func newUninstallAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall [agent-id]",
		Short: "Uninstall agent configuration",
		Long:  `Remove an agent's configuration from the system.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runUninstallAgent,
	}

	return cmd
}

func runUninstallAgent(cmd *cobra.Command, args []string) error {
	agentID := args[0]

	agent := DetectCLIByName(agentID)
	if agent == nil {
		return fmt.Errorf("unknown agent: %s", agentID)
	}

	if err := agent.Uninstall(); err != nil {
		return fmt.Errorf("failed to uninstall %s: %w", agent.Name(), err)
	}

	fmt.Printf("✓ %s configuration uninstalled successfully\n", agent.Name())
	return nil
}

// newInstallAllCmd creates the install-all command
func newInstallAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all-agents",
		Short: "Install configurations for all available agents",
		Long:  `Detect available agents and install configurations for all of them.`,
		RunE:  runInstallAll,
	}

	cmd.Flags().String("model", "", "Model to use for all agents")

	return cmd
}

func runInstallAll(cmd *cobra.Command, args []string) error {
	model := getStringFlag(cmd, "model")
	agents := GetInstalledAgents()

	if len(agents) == 0 {
		fmt.Println("No agent CLIs detected. Checking all supported agents...")

		// Install for all agents even if CLI is not available
		agents = ListAgents()
	}

	installed := 0
	failed := 0

	for _, agent := range agents {
		cfg := &AgentConfig{Model: model}

		if err := agent.Install(cfg); err != nil {
			logger.Warn("Failed to install", "agent", agent.Name(), "error", err)
			failed++
			continue
		}

		installed++
		fmt.Printf("✓ %s configuration installed\n", agent.Name())
	}

	fmt.Printf("\nInstallation complete: %d succeeded, %d failed\n", installed, failed)

	if failed > 0 {
		return fmt.Errorf("%d agents failed to install", failed)
	}

	return nil
}

// getStringFlag safely gets a string flag value
func getStringFlag(cmd *cobra.Command, name string) string {
	val, _ := cmd.Flags().GetString(name)
	return val
}
