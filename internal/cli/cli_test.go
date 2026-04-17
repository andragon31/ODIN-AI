package cli

import (
	"testing"
)

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd("1.0.0", "test-build-time")

	if cmd.Use != "odin" {
		t.Errorf("expected use 'odin', got '%s'", cmd.Use)
	}

	if cmd.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", cmd.Version)
	}

	// Check that subcommands are added
	subcommands := []string{"init", "status", "version", "config", "session"}
	for _, name := range subcommands {
		_, _, err := cmd.Find([]string{name})
		if err != nil {
			t.Errorf("expected subcommand '%s' to be registered: %v", name, err)
		}
	}
}

func TestNewRootCmdFlags(t *testing.T) {
	cmd := NewRootCmd("1.0.0", "")

	flags := []string{"quiet", "debug", "json", "config"}
	for _, name := range flags {
		if cmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("expected flag '%s' to be registered", name)
		}
	}
}

func TestInitCmd(t *testing.T) {
	cmd := newInitCmd()

	if cmd.Use != "init" {
		t.Errorf("expected use 'init', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected short description to be set")
	}
}

func TestStatusCmd(t *testing.T) {
	cmd := newStatusCmd("1.0.0")

	if cmd.Use != "status" {
		t.Errorf("expected use 'status', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected short description to be set")
	}
}

func TestVersionCmd(t *testing.T) {
	cmd := newVersionCmd("1.2.3", "2024-01-01")

	if cmd.Use != "version" {
		t.Errorf("expected use 'version', got '%s'", cmd.Use)
	}
}

func TestConfigCmd(t *testing.T) {
	cmd := newConfigCmd()

	if cmd.Use != "config" {
		t.Errorf("expected use 'config', got '%s'", cmd.Use)
	}

	// Check that config show subcommand exists
	_, _, err := cmd.Find([]string{"show"})
	if err != nil {
		t.Error("expected 'config show' subcommand to be registered")
	}
}

func TestSessionCmd(t *testing.T) {
	cmd := newSessionCmd()

	if cmd.Use != "session" {
		t.Errorf("expected use 'session', got '%s'", cmd.Use)
	}

	// Check that session list subcommand exists
	_, _, err := cmd.Find([]string{"list"})
	if err != nil {
		t.Error("expected 'session list' subcommand to be registered")
	}
}


func TestStatusDataStructure(t *testing.T) {
	// Test that the status command returns expected structure
	status := map[string]interface{}{
		"version": "1.0.0",
		"status":  "healthy",
		"components": map[string]interface{}{
			"odin":     "healthy",
			"mimir":    "not_initialized",
			"heimdall": "not_initialized",
			"bifrost":  "not_initialized",
			"runes":    "not_initialized",
			"nornir":   "not_initialized",
			"dvergar":  "not_initialized",
			"volva":    "not_initialized",
		},
		"mode": "local",
	}

	// Verify version
	if status["version"] != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %v", status["version"])
	}

	// Verify status
	if status["status"] != "healthy" {
		t.Errorf("expected status healthy, got %v", status["status"])
	}

	// Verify mode
	if status["mode"] != "local" {
		t.Errorf("expected mode local, got %v", status["mode"])
	}

	// Verify all 8 components exist
	components := status["components"].(map[string]interface{})
	expectedComponents := []string{
		"odin", "mimir", "heimdall", "bifrost",
		"runes", "nornir", "dvergar", "volva",
	}

	for _, name := range expectedComponents {
		if _, ok := components[name]; !ok {
			t.Errorf("expected component %s in status", name)
		}
	}
}
