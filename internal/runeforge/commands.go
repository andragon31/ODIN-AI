package runeforge

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/internal/router"
	"github.com/odin-ai/odin/internal/skills"
)

// Commands returns all RuneForge CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rune",
		Short: "RuneForge - Generate new runes via local Router",
		Long: `RuneForge generates new skills using ODIN's local Router.
Unlike external AI CLIs, RuneForge works offline with Ollama.`,
	}

	cmd.AddCommand(
		newForgeCmd(),
		newValidateForgeCmd(),
	)

	return cmd
}

// newForgeCmd creates the forge command
func newForgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forge [name]",
		Short: "Generate a new rune",
		Long: `Generate a new Rune skill using the local Router.
Example:
  odin rune forge "branch-pr" --description "PR workflow" --tags git,pr`,
		RunE: runForge,
	}

	cmd.Flags().String("description", "", "Rune description")
	cmd.Flags().StringSlice("tags", nil, "Comma-separated tags")
	cmd.Flags().String("model", "ollama:deepseek-coder", "Model to use")
	cmd.Flags().String("from-example", "", "Adapt from an existing rune file")
	cmd.Flags().String("adapt-for", "odin", "Platform to adapt for")
	cmd.Flags().Duration("timeout", 2*time.Minute, "Generation timeout")

	return cmd
}

func runForge(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	var name string
	if len(args) > 0 {
		name = args[0]
	}

	description, _ := cmd.Flags().GetString("description")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	model, _ := cmd.Flags().GetString("model")
	fromExample, _ := cmd.Flags().GetString("from-example")
	adaptFor, _ := cmd.Flags().GetString("adapt-for")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	if name == "" && fromExample == "" {
		return fmt.Errorf("either name or --from-example is required")
	}

	// Initialize router
	r, err := initRouter()
	if err != nil {
		return fmt.Errorf("failed to initialize router: %w", err)
	}

	forge := NewRuneForge(r)

	var result *ForgeResult

	if fromExample != "" {
		// Generate from example
		result, err = forge.GenerateFromExample(ctx, fromExample, adaptFor)
	} else {
		// Generate new rune
		req := ForgeRequest{
			Name:        name,
			Description: description,
			Tags:        tags,
			Model:       model,
		}
		result, err = forge.GenerateWithTimeout(ctx, req, timeout)
	}

	if err != nil {
		return fmt.Errorf("forge failed: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	// Display result
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)

	if result.Valid {
		fmt.Fprintf(w, "✓ Rune generated successfully\n")
		fmt.Fprintf(w, "Name:\t%s\n", result.Rune.Name)
		fmt.Fprintf(w, "Version:\t%s\n", result.Rune.Version)
		fmt.Fprintf(w, "Description:\t%s\n", result.Rune.Description)
		fmt.Fprintf(w, "Tags:\t%s\n", joinTags(result.Rune.Tags))
		fmt.Fprintf(w, "Type:\t%s\n", result.Rune.Execution.Type)
	} else {
		fmt.Fprintf(w, "✗ Rune generation failed\n")
		for _, err := range result.Errors {
			fmt.Fprintf(w, "Error:\t%s\n", err)
		}
	}
	w.Flush()

	return nil
}

// newValidateForgeCmd creates the validate command for forged runes
func newValidateForgeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [rune-file]",
		Short: "Validate a rune file",
		Long:  `Validate a rune YAML file against the schema.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runValidateForge,
	}
}

func runValidateForge(cmd *cobra.Command, args []string) error {
	runePath := args[0]

	rune, validationResult, err := validateRuneFile(runePath)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{
			"valid":    validationResult.Valid,
			"errors":   validationResult.Errors,
			"warnings": validationResult.Warns,
			"rune":     rune,
		})
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	if validationResult.Valid {
		fmt.Fprintf(w, "✓ Rune %s@%s is valid\n", rune.Name, rune.Version)
	} else {
		fmt.Fprintf(w, "✗ Rune is invalid\n")
		for _, err := range validationResult.Errors {
			fmt.Fprintf(w, "Error:\t%s\n", err)
		}
	}
	w.Flush()

	if !validationResult.Valid {
		return fmt.Errorf("validation failed")
	}

	return nil
}

// initRouter initializes the router for rune generation
func initRouter() (*router.Router, error) {
	// Try to create a basic router with Ollama
	ollamaProvider := router.NewOllamaProvider(router.OllamaConfig{
		Enabled:  true,
		Endpoint: "http://localhost:11434",
	})

	r, err := router.NewRouter([]router.Provider{ollamaProvider}, "ollama")
	if err != nil {
		return nil, err
	}

	return r, nil
}

// validateRuneFile validates a rune file
func validateRuneFile(path string) (*skills.Rune, *validationResult, error) {
	rune, valResult, err := skills.ValidateFile(path)
	if err != nil {
		return nil, &validationResult{Valid: false, Errors: []string{err.Error()}, Warns: []string{}}, err
	}
	return rune, &validationResult{
		Valid:  valResult.Valid,
		Errors: valResult.Errors,
		Warns:  valResult.Warns,
	}, nil
}

// validationResult holds validation results
type validationResult struct {
	Valid  bool
	Errors []string
	Warns  []string
}

// joinTags joins tags into a comma-separated string
func joinTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	result := tags[0]
	for i := 1; i < len(tags); i++ {
		result += ", " + tags[i]
	}
	return result
}
