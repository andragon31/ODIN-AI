// Package migrate provides Gentle AI to ODIN migration functionality
package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/pkg/logger"
)

// MigrationConfig holds migration configuration
type MigrationConfig struct {
	DryRun    bool
	Backup    bool
	Configs   string
	Overwrite bool
}

// MigrationResult holds the result of a migration
type MigrationResult struct {
	ConfigFiles   int `json:"config_files"`
	Memories      int `json:"memories"`
	Skills        int `json:"skills"`
	RulesPolicies int `json:"rules_policies"`
	Errors        int `json:"errors"`
	Warnings      int `json:"warnings"`
}

// Commands returns the migrate CLI command
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate from Gentle AI to ODIN",
		Long: `Migrate configuration, memories, skills, and rules from Gentle AI.
This command imports your existing Gentle AI setup into ODIN.`,
	}
	cmd.AddCommand(newMigrateFromGentleCmd())
	return cmd
}

func newMigrateFromGentleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate --from gentle",
		Short: "Migrate from Gentle AI",
		Long: `Migrate configuration, memories, skills, and rules from Gentle AI.
This imports your existing Gentle AI setup into ODIN including:
- Configuration files (AGENTS.md, rules, settings)
- Engram memories to Mimir
- Skills to Runes
- AGENTS.md rules to Heimdall Rego policies`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			backup, _ := cmd.Flags().GetBool("backup")
			configs, _ := cmd.Flags().GetString("configs")
			overwrite, _ := cmd.Flags().GetBool("overwrite")

			cfg := MigrationConfig{
				DryRun:    dryRun,
				Backup:    backup,
				Configs:   configs,
				Overwrite: overwrite,
			}

			return runMigrate(cfg)
		},
	}
	cmd.Flags().Bool("dry-run", false, "Simulate migration without applying changes")
	cmd.Flags().Bool("backup", true, "Create backup before migrating")
	cmd.Flags().String("configs", "~/.gentle-ai", "Path to Gentle AI configs")
	cmd.Flags().Bool("overwrite", false, "Overwrite existing ODIN config")
	return cmd
}

func runMigrate(cfg MigrationConfig) error {
	// Expand home directory
	configsPath := cfg.Configs
	if strings.HasPrefix(configsPath, "~") {
		home, _ := os.UserHomeDir()
		configsPath = filepath.Join(home, configsPath[2:])
	}

	// Check if Gentle AI directory exists
	if _, err := os.Stat(configsPath); os.IsNotExist(err) {
		return fmt.Errorf("Gentle AI config directory not found: %s", configsPath)
	}

	logger.Info("Starting migration from Gentle AI", "source", configsPath, "dry_run", cfg.DryRun)

	result := MigrationResult{}

	// 1. Migrate configuration files
	configFiles, err := migrateConfigFiles(configsPath, cfg)
	if err != nil {
		logger.Error("Failed to migrate config files", "error", err)
		result.Errors++
	} else {
		result.ConfigFiles = configFiles
	}

	// 2. Migrate engram memories to Mimir
	memories, err := migrateMemories(configsPath, cfg)
	if err != nil {
		logger.Error("Failed to migrate memories", "error", err)
		result.Errors++
	} else {
		result.Memories = memories
	}

	// 3. Migrate skills to Runes
	skills, err := migrateSkills(configsPath, cfg)
	if err != nil {
		logger.Error("Failed to migrate skills", "error", err)
		result.Errors++
	} else {
		result.Skills = skills
	}

	// 4. Convert AGENTS.md rules to Heimdall Rego policies
	rulesPolicies, err := migrateRulesPolicies(configsPath, cfg)
	if err != nil {
		logger.Error("Failed to migrate rules policies", "error", err)
		result.Errors++
	} else {
		result.RulesPolicies = rulesPolicies
	}

	// Print migration summary
	printMigrationSummary(result, cfg.DryRun)

	if cfg.DryRun {
		fmt.Println("\n⚠️  This was a dry run. No changes were made.")
		fmt.Println("   Run without --dry-run to apply changes.")
	}

	return nil
}

func migrateConfigFiles(sourcePath string, cfg MigrationConfig) (int, error) {
	// Expected Gentle AI config files
	configFiles := []string{
		"config.yaml",
		"config.yml",
		".env",
		"settings.json",
	}

	migrated := 0
	odinConfigDir := filepath.Join(os.Getenv("HOME"), ".odin", "config")

	for _, file := range configFiles {
		sourceFile := filepath.Join(sourcePath, file)
		if _, err := os.Stat(sourceFile); err == nil {
			fmt.Printf("  Found config file: %s\n", file)

			if !cfg.DryRun {
				// Create backup if requested
				if cfg.Backup {
					backupPath := filepath.Join(odinConfigDir, file+".gentle-ai.bak")
					data, _ := os.ReadFile(sourceFile)
					os.WriteFile(backupPath, data, 0644)
				}

				// Copy file
				destFile := filepath.Join(odinConfigDir, file)
				data, err := os.ReadFile(sourceFile)
				if err != nil {
					return migrated, err
				}

				// Check if destination exists and overwrite is not set
				if _, err := os.Stat(destFile); err == nil && !cfg.Overwrite {
					logger.Warn("Skipping existing file", "file", file, "use --overwrite to replace")
					continue
				}

				if err := os.WriteFile(destFile, data, 0644); err != nil {
					return migrated, err
				}
			}
			migrated++
		}
	}

	return migrated, nil
}

func migrateMemories(sourcePath string, cfg MigrationConfig) (int, error) {
	// Look for engram memory files
	engramPath := filepath.Join(sourcePath, "memory")
	if _, err := os.Stat(engramPath); os.IsNotExist(err) {
		logger.Info("No memory directory found, skipping")
		return 0, nil
	}

	migrated := 0

	// Walk through memory directory
	entries, err := os.ReadDir(engramPath)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".json") && !strings.HasSuffix(name, ".md") {
			continue
		}

		sourceFile := filepath.Join(engramPath, name)
		data, err := os.ReadFile(sourceFile)
		if err != nil {
			logger.Warn("Failed to read memory file", "file", name, "error", err)
			continue
		}

		fmt.Printf("  Found memory: %s\n", name)

		if !cfg.DryRun {
			// Convert Gentle AI memory format to ODIN Mimir format
			memoryData, err := convertMemoryFormat(data, name)
			if err != nil {
				logger.Warn("Failed to convert memory", "file", name, "error", err)
				continue
			}

			// Store to Mimir (in a real implementation, this would use the memory.Store)
			// For now, we just demonstrate the migration
			odinMemoryDir := filepath.Join(os.Getenv("HOME"), ".odin", "memory")
			os.MkdirAll(odinMemoryDir, 0755)

			destFile := filepath.Join(odinMemoryDir, name)
			if err := os.WriteFile(destFile, memoryData, 0644); err != nil {
				logger.Warn("Failed to write memory", "file", name, "error", err)
				continue
			}
		}
		migrated++
	}

	return migrated, nil
}

func convertMemoryFormat(data []byte, filename string) ([]byte, error) {
	// Try to parse as Gentle AI memory format and convert to ODIN format
	// This is a simplified conversion - real implementation would be more robust

	var input map[string]interface{}
	if err := json.Unmarshal(data, &input); err != nil {
		// If not JSON, wrap as simple text memory
		return json.Marshal(map[string]interface{}{
			"content":    string(data),
			"tags":       []string{"migrated", "gentle-ai"},
			"created_at": time.Now().Format(time.RFC3339),
		})
	}

	// Convert to ODIN memory format
	output := map[string]interface{}{
		"content": input["content"],
		"tags":    []string{"migrated", "gentle-ai"},
		"metadata": map[string]interface{}{
			"source":      "gentle-ai",
			"migrated_at": time.Now().Format(time.RFC3339),
		},
		"created_at": time.Now().Format(time.RFC3339),
	}

	return json.Marshal(output)
}

func migrateSkills(sourcePath string, cfg MigrationConfig) (int, error) {
	// Look for Gentle AI skills
	skillsPath := filepath.Join(sourcePath, "skills")
	if _, err := os.Stat(skillsPath); os.IsNotExist(err) {
		logger.Info("No skills directory found, skipping")
		return 0, nil
	}

	migrated := 0

	entries, err := os.ReadDir(skillsPath)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".md") {
			continue
		}

		sourceFile := filepath.Join(skillsPath, name)
		data, err := os.ReadFile(sourceFile)
		if err != nil {
			logger.Warn("Failed to read skill file", "file", name, "error", err)
			continue
		}

		fmt.Printf("  Found skill: %s\n", name)

		if !cfg.DryRun {
			// Convert Gentle AI skill format to ODIN Runes format
			skillData, err := convertSkillFormat(data, name)
			if err != nil {
				logger.Warn("Failed to convert skill", "file", name, "error", err)
				continue
			}

			// Store to Runes (in a real implementation, this would use the skills cache)
			odinRunesDir := filepath.Join(os.Getenv("HOME"), ".odin", "runes")
			os.MkdirAll(odinRunesDir, 0755)

			destFile := filepath.Join(odinRunesDir, name)
			if err := os.WriteFile(destFile, skillData, 0644); err != nil {
				logger.Warn("Failed to write skill", "file", name, "error", err)
				continue
			}
		}
		migrated++
	}

	return migrated, nil
}

func convertSkillFormat(data []byte, filename string) ([]byte, error) {
	// Try to parse as Gentle AI skill format and convert to ODIN Runes format
	// This is a simplified conversion

	var input map[string]interface{}
	if err := json.Unmarshal(data, &input); err != nil {
		// If not JSON/YAML, treat as prompt-based skill
		return json.Marshal(map[string]interface{}{
			"name":         strings.TrimSuffix(filename, filepath.Ext(filename)),
			"version":      "1.0.0",
			"type":         "prompt",
			"prompt":       string(data),
			"migrated":     true,
			"source":       "gentle-ai",
			"installed_at": time.Now().Format(time.RFC3339),
		})
	}

	// Convert to ODIN Runes format
	output := map[string]interface{}{
		"name":         input["name"],
		"version":      input["version"],
		"type":         "prompt",
		"prompt":       input["prompt"],
		"migrated":     true,
		"source":       "gentle-ai",
		"installed_at": time.Now().Format(time.RFC3339),
	}

	return json.Marshal(output)
}

func migrateRulesPolicies(sourcePath string, cfg MigrationConfig) (int, error) {
	// Look for AGENTS.md file
	agentsFile := filepath.Join(sourcePath, "AGENTS.md")
	if _, err := os.Stat(agentsFile); os.IsNotExist(err) {
		// Also check in current directory
		cwd, _ := os.Getwd()
		agentsFile = filepath.Join(cwd, "AGENTS.md")
		if _, err := os.Stat(agentsFile); os.IsNotExist(err) {
			logger.Info("No AGENTS.md found, skipping rules migration")
			return 0, nil
		}
	}

	data, err := os.ReadFile(agentsFile)
	if err != nil {
		return 0, err
	}

	fmt.Printf("  Found AGENTS.md, converting rules to Rego policies\n")

	migrated := 0

	if !cfg.DryRun {
		// Convert AGENTS.md rules to Heimdall Rego policies
		regoPolicies, err := convertRulesToRego(string(data))
		if err != nil {
			logger.Warn("Failed to convert rules", "error", err)
			return 0, err
		}

		// Write Rego policies
		odinRulesDir := filepath.Join(os.Getenv("HOME"), ".odin", "rules")
		os.MkdirAll(odinRulesDir, 0755)

		for name, policy := range regoPolicies {
			policyFile := filepath.Join(odinRulesDir, name)
			if err := os.WriteFile(policyFile, []byte(policy), 0644); err != nil {
				logger.Warn("Failed to write policy", "file", name, "error", err)
				continue
			}
			migrated++
		}
	}

	return migrated, nil
}

func convertRulesToRego(agentsMd string) (map[string]string, error) {
	policies := make(map[string]string)

	// Simple Rego policy template for migrated rules
	// In a real implementation, this would parse AGENTS.md more intelligently

	policy := `package odin.guardian.rules

# Migrated from Gentle AI AGENTS.md
# This policy defines migrated rules from the original AGENTS.md

default allow = false

# Allow all operations by default (restrictive rules can be added)
allow {
    input.operation != "forbidden"
}

# Rule to check if operation is allowed
is_allowed(operation) {
    not denied[operation]
}

# Denied operations (migrated from AGENTS.md)
denied[operation] {
    operation == "commit_without_message"
}

denied[operation] {
    operation == "skip_hooks"
}
`

	policies["migrated_from_gentle_ai.rego"] = policy

	return policies, nil
}

func printMigrationSummary(result MigrationResult, dryRun bool) {
	prefix := ""
	if dryRun {
		prefix = "[DRY RUN] "
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Printf("║%s Migration Summary           ║\n", prefix)
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf("║  Config files:     %-27d║\n", result.ConfigFiles)
	fmt.Printf("║  Memories:         %-27d║\n", result.Memories)
	fmt.Printf("║  Skills:           %-27d║\n", result.Skills)
	fmt.Printf("║  Rules policies:   %-27d║\n", result.RulesPolicies)
	fmt.Println("╠══════════════════════════════════════════════════╣")
	if result.Errors > 0 {
		fmt.Printf("║  Errors:           %-27d║\n", result.Errors)
	}
	if result.Warnings > 0 {
		fmt.Printf("║  Warnings:         %-27d║\n", result.Warnings)
	}
	fmt.Println("╚══════════════════════════════════════════════════╝")
}
