package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", cfg.Version)
	}

	if cfg.Mode != "local" {
		t.Errorf("expected mode local, got %s", cfg.Mode)
	}

	if cfg.HomeDir == "" {
		t.Error("expected home dir to be set")
	}

	// Check Memory config
	if cfg.Memory.Engine != "sqlite-vss" {
		t.Errorf("expected memory engine sqlite-vss, got %s", cfg.Memory.Engine)
	}
	if !cfg.Memory.Encryption {
		t.Error("expected memory encryption to be enabled")
	}
	if len(cfg.Memory.Pruning.KeepTags) == 0 {
		t.Error("expected pruning keep tags to be set")
	}

	// Check Guardian config
	if cfg.Guardian.PolicyEngine != "opa" {
		t.Errorf("expected guardian policy engine opa, got %s", cfg.Guardian.PolicyEngine)
	}
	if !cfg.Guardian.BlockOnCrit {
		t.Error("expected guardian to block on critical")
	}

	// Check Router config
	if cfg.Router.Default != "ollama-local" {
		t.Errorf("expected router default ollama-local, got %s", cfg.Router.Default)
	}
	if len(cfg.Router.Fallback) == 0 {
		t.Error("expected router fallback to be set")
	}

	// Check Session config
	if cfg.Session.MaxSessions != 10 {
		t.Errorf("expected max sessions 10, got %d", cfg.Session.MaxSessions)
	}
}

func TestMemoryConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Memory.Engine != "sqlite-vss" {
		t.Errorf("expected sqlite-vss, got %s", cfg.Memory.Engine)
	}

	expectedTags := []string{"arch", "spec", "security"}
	if len(cfg.Memory.Pruning.KeepTags) != len(expectedTags) {
		t.Errorf("expected %d keep tags, got %d", len(expectedTags), len(cfg.Memory.Pruning.KeepTags))
	}
}

func TestSyncConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Sync.Backend != "git" {
		t.Errorf("expected git backend, got %s", cfg.Sync.Backend)
	}

	if cfg.Sync.AutoPush {
		t.Error("expected auto push to be disabled by default")
	}

	if cfg.Sync.GPGSign {
		t.Error("expected GPG sign to be disabled by default")
	}
}

func TestGuardianConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Guardian.PolicyEngine != "opa" {
		t.Errorf("expected opa policy engine, got %s", cfg.Guardian.PolicyEngine)
	}

	if !cfg.Guardian.SAST.Enabled {
		t.Error("expected SAST to be enabled")
	}

	if len(cfg.Guardian.SAST.Tools) != 2 {
		t.Errorf("expected 2 SAST tools, got %d", len(cfg.Guardian.SAST.Tools))
	}
}

func TestRouterConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Router.Default != "ollama-local" {
		t.Errorf("expected ollama-local default, got %s", cfg.Router.Default)
	}

	expectedFallback := []string{"openrouter", "anthropic"}
	if len(cfg.Router.Fallback) != len(expectedFallback) {
		t.Errorf("expected %d fallback providers, got %d", len(expectedFallback), len(cfg.Router.Fallback))
	}

	if cfg.Router.CostCapDay != 0.0 {
		t.Errorf("expected 0 cost cap, got %f", cfg.Router.CostCapDay)
	}
}

func TestObservConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Observability.MetricsPort != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Observability.MetricsPort)
	}

	if cfg.Observability.LogLevel != "info" {
		t.Errorf("expected info log level, got %s", cfg.Observability.LogLevel)
	}
}

func TestPluginsConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Plugins.Sandbox {
		t.Error("expected sandbox to be enabled")
	}

	if cfg.Plugins.AutoUpdate {
		t.Error("expected auto update to be disabled")
	}
}

func TestSessionConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Session.MaxSessions != 10 {
		t.Errorf("expected max sessions 10, got %d", cfg.Session.MaxSessions)
	}

	if cfg.Session.SnapshotInterval != "5m" {
		t.Errorf("expected snapshot interval 5m, got %s", cfg.Session.SnapshotInterval)
	}
}

func TestLoadWithDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", cfg.Version)
	}

	if cfg.Mode != "local" {
		t.Errorf("expected mode local, got %s", cfg.Mode)
	}
}

func TestLoadWithCustomPath(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
version: "1.5"
mode: "docker"
memory:
  engine: "pgvector"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Version != "1.5" {
		t.Errorf("expected version 1.5, got %s", cfg.Version)
	}

	if cfg.Mode != "docker" {
		t.Errorf("expected mode docker, got %s", cfg.Mode)
	}
}

func TestLoadInvalidPath(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error loading nonexistent config")
	}
}

func TestEnsureDirs(t *testing.T) {
	cfg := DefaultConfig()

	// Use temp directory for testing
	tmpDir := t.TempDir()
	cfg.HomeDir = tmpDir
	cfg.Memory.Path = filepath.Join(tmpDir, ".odin", "memory.db")
	cfg.Guardian.RulesPath = filepath.Join(tmpDir, ".odin", "rules")
	cfg.Observability.LogPath = filepath.Join(tmpDir, ".odin", "logs")
	cfg.Plugins.Allowed = []string{filepath.Join(tmpDir, ".odin", "plugins")}
	cfg.Themes.Path = filepath.Join(tmpDir, ".odin", "themes")
	cfg.Session.Path = filepath.Join(tmpDir, ".odin", "sessions")

	err := cfg.EnsureDirs()
	if err != nil {
		t.Fatalf("unexpected error ensuring dirs: %v", err)
	}

	// Verify directories were created
	dirs := []string{
		filepath.Dir(cfg.Memory.Path),
		cfg.Guardian.RulesPath,
		cfg.Observability.LogPath,
		cfg.Plugins.Allowed[0],
		cfg.Themes.Path,
		cfg.Session.Path,
	}

	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("directory %s was not created: %v", dir, err)
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", dir)
		}
	}
}

func TestHomeDir(t *testing.T) {
	cfg := DefaultConfig()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home directory")
	}

	if cfg.HomeDir != homeDir {
		t.Errorf("expected home dir %s, got %s", homeDir, cfg.HomeDir)
	}
}

func TestMemoryPathInHomeDir(t *testing.T) {
	cfg := DefaultConfig()

	expectedPath := filepath.Join(cfg.HomeDir, ".odin", "memory.db")
	if cfg.Memory.Path != expectedPath {
		t.Errorf("expected memory path %s, got %s", expectedPath, cfg.Memory.Path)
	}
}

func TestGuardianRulesPathInHomeDir(t *testing.T) {
	cfg := DefaultConfig()

	expectedPath := filepath.Join(cfg.HomeDir, ".odin", "rules")
	if cfg.Guardian.RulesPath != expectedPath {
		t.Errorf("expected rules path %s, got %s", expectedPath, cfg.Guardian.RulesPath)
	}
}
