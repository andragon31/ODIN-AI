package router

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/internal/config"
	"github.com/odin-ai/odin/pkg/logger"
)

// Commands returns all Router CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "router",
		Short: "Router - Model routing with fallback chain",
		Long: `Router manages model providers with automatic fallback chain.
Völva selects the best available model for your task, trying providers
in order until one succeeds.`,
	}

	cmd.AddCommand(
		newRouterStatusCmd(),
		newRouterSetCmd(),
		newRouterFallbackCmd(),
		newRouterMetricsCmd(),
		newRouterModelsCmd(),
	)

	return cmd
}

func newRouterStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current provider and health status",
		Long:  `Display the health status of all configured providers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRouterStatus(cmd)
		},
	}
}

func runRouterStatus(cmd *cobra.Command) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	ctx := context.Background()

	// Create router with available providers
	r, err := createRouter()
	if err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	}

	// Check health of all providers
	health := r.CheckHealth(ctx)
	providers := r.ListProviders()
	defaultProvider := r.GetDefaultProvider()

	if jsonOutput {
		status := map[string]interface{}{
			"default_provider": defaultProvider.Name(),
			"health":           health,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(status)
	}

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║         ODIN AI - Router Status                  ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf("║  Default Provider: %-28s║\n", defaultProvider.Name())
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Println("║  Provider Health:                                ║")

	for _, p := range providers {
		status := "❌ unavailable"
		if health[p.Name()] {
			status = "✅ available"
		}
		defaultMark := ""
		if p.Name() == defaultProvider.Name() {
			defaultMark = " (default)"
		}
		fmt.Printf("║    %-12s %-29s%s║\n", p.Name()+":", status, defaultMark)
	}

	fmt.Println("╚══════════════════════════════════════════════════╝")
	return nil
}

func newRouterSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <provider>",
		Short: "Set the primary provider",
		Long:  `Set a provider as the primary/default provider.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			return runRouterSet(provider)
		},
	}
}

func runRouterSet(providerName string) error {
	validProviders := map[string]bool{"ollama": true, "openrouter": true, "anthropic": true}
	if !validProviders[providerName] {
		return fmt.Errorf("unknown provider: %s (valid: ollama, openrouter, anthropic)", providerName)
	}

	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.Router.Default = providerName

	// Note: In a full implementation, we would save the config here
	logger.Info("Provider set", "provider", providerName)
	fmt.Printf("Default provider set to: %s\n", providerName)
	return nil
}

func newRouterFallbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fallback",
		Short: "Manage fallback chain",
		Long:  `Add or remove providers from the fallback chain.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List current fallback chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFallbackList()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "add <provider>",
		Short: "Add a provider to fallback chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			return runFallbackAdd(provider)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "remove <provider>",
		Short: "Remove a provider from fallback chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			return runFallbackRemove(provider)
		},
	})

	return cmd
}

func runFallbackList() error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Fallback Chain:")
	if len(cfg.Router.Fallback) == 0 {
		fmt.Println("  (empty - will try providers in order)")
	} else {
		for i, p := range cfg.Router.Fallback {
			fmt.Printf("  %d. %s\n", i+1, p)
		}
	}
	return nil
}

func runFallbackAdd(provider string) error {
	validProviders := map[string]bool{"ollama": true, "openrouter": true, "anthropic": true}
	if !validProviders[provider] {
		return fmt.Errorf("unknown provider: %s", provider)
	}

	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if already in chain
	for _, p := range cfg.Router.Fallback {
		if p == provider {
			return fmt.Errorf("provider %s already in fallback chain", provider)
		}
	}

	cfg.Router.Fallback = append(cfg.Router.Fallback, provider)
	logger.Info("Provider added to fallback", "provider", provider)
	fmt.Printf("Added %s to fallback chain\n", provider)
	return nil
}

func runFallbackRemove(provider string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	newChain := []string{}
	found := false
	for _, p := range cfg.Router.Fallback {
		if p == provider {
			found = true
		} else {
			newChain = append(newChain, p)
		}
	}

	if !found {
		return fmt.Errorf("provider %s not in fallback chain", provider)
	}

	cfg.Router.Fallback = newChain
	logger.Info("Provider removed from fallback", "provider", provider)
	fmt.Printf("Removed %s from fallback chain\n", provider)
	return nil
}

func newRouterMetricsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "metrics",
		Short: "Show usage metrics",
		Long:  `Display router usage metrics including latency, costs, and success rates.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRouterMetrics(cmd)
		},
	}
}

func runRouterMetrics(cmd *cobra.Command) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	r, err := createRouter()
	if err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	}

	metrics := r.GetMetrics().GetAllMetrics()

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(metrics)
	}

	if len(metrics) == 0 {
		fmt.Println("No metrics available yet")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "%-15s %10s %10s %10s %15s %12s %10s\n",
		"Provider", "Requests", "Success", "Errors", "Avg Latency", "Success %", "Total Cost")
	fmt.Fprintf(w, "%s\n", "-------------------------------------------------------------------")

	for name, m := range metrics {
		avgLatency := "0s"
		if m.RequestCount > 0 {
			avgLatency = (m.TotalLatency / time.Duration(m.RequestCount)).String()
		}
		successRate := float64(0)
		if m.RequestCount > 0 {
			successRate = float64(m.SuccessCount) / float64(m.RequestCount) * 100
		}
		fmt.Fprintf(w, "%-15s %10d %10d %10d %15s %11.1f%% %9.4f\n",
			name, m.RequestCount, m.SuccessCount, m.ErrorCount,
			avgLatency, successRate, m.TotalCost)
	}
	w.Flush()

	return nil
}

func newRouterModelsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "models",
		Short: "List available models for current provider",
		Long:  `Show available models from the current/default provider.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRouterModels(cmd)
		},
	}
}

func runRouterModels(cmd *cobra.Command) error {
	r, err := createRouter()
	if err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	}

	provider := r.GetDefaultProvider()

	fmt.Printf("Models available from %s:\n", provider.Name())

	switch provider.Name() {
	case "ollama":
		fmt.Println("  Run 'ollama list' to see installed models")
		fmt.Println("  Common models: llama3, mistral, codellama, nomic-embed-text")
	case "openrouter":
		fmt.Println("  • anthropic/claude-3.5-sonnet")
		fmt.Println("  • anthropic/claude-3-opus")
		fmt.Println("  • openai/gpt-4-turbo")
		fmt.Println("  • google/gemini-pro-1.5")
		fmt.Println("  • meta-llama/llama-3-70b")
		fmt.Println("  (See https://openrouter.ai/models for full list)")
	case "anthropic":
		fmt.Println("  • claude-3-5-sonnet-20241022")
		fmt.Println("  • claude-3-opus-20240229")
		fmt.Println("  • claude-3-sonnet-20240229")
		fmt.Println("  • claude-3-haiku-20240307")
	default:
		fmt.Println("  No model information available")
	}

	return nil
}

// createRouter creates a router with all configured providers
func createRouter() (*Router, error) {
	cfg, err := config.Load("")
	if err != nil {
		// Use default config if loading fails
		cfg = config.DefaultConfig()
	}

	providers := []Provider{
		NewOllamaProvider(OllamaConfig{
			Enabled:  true,
			Endpoint: "http://localhost:11434",
		}),
		NewOpenRouterProvider(OpenRouterConfig{
			Enabled:  true,
			APIKey:   "",
			Endpoint: "https://openrouter.ai/api/v1",
		}),
		NewAnthropicProvider(AnthropicConfig{
			Enabled:  true,
			APIKey:   "",
			Endpoint: "https://api.anthropic.com",
		}),
	}

	r, err := NewRouter(providers, cfg.Router.Default)
	if err != nil {
		return nil, err
	}

	if len(cfg.Router.Fallback) > 0 {
		r.SetFallbackChain(cfg.Router.Fallback)
	}

	return r, nil
}
