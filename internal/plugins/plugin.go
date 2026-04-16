// Package plugins provides the WASM plugin runtime for ODIN
// Dvergar are the Norse dwarves who craft powerful artifacts
// This package implements the plugin interface and WASM runtime
package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Plugin interface that all ODIN plugins must implement
type Plugin interface {
	Metadata() PluginMetadata
	Init(ctx context.Context, config json.RawMessage) error
	Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
	Health(ctx context.Context) error
}

// PluginMetadata contains plugin information
type PluginMetadata struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"` // "fs:readonly", "net:disallow", etc.
}

// Manifest represents a plugin manifest file
type Manifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	EntryPoint  string   `json:"entry_point"` // Path to WASM binary
	Permissions []string `json:"permissions"` // Required permissions
}

// Validate checks if the manifest is valid
func (m *Manifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if m.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	if m.EntryPoint == "" {
		return fmt.Errorf("plugin entry point is required")
	}
	return nil
}

// PluginInstance represents a loaded plugin instance
type PluginInstance struct {
	Metadata PluginMetadata
	plugin   Plugin
	runtime  *Runtime
	manifest *Manifest
}

// Runtime configuration
type RuntimeConfig struct {
	PluginsDir string
	Sandbox    bool
}

// DefaultRuntimeConfig returns the default runtime configuration
func DefaultRuntimeConfig() *RuntimeConfig {
	homeDir, _ := os.UserHomeDir()
	return &RuntimeConfig{
		PluginsDir: filepath.Join(homeDir, ".odin", "plugins"),
		Sandbox:    true,
	}
}

// ValidatePluginManifest validates a plugin manifest file
func ValidatePluginManifest(manifestPath string) (*Manifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	// Check that entry point exists
	manifestDir := filepath.Dir(manifestPath)
	entryPath := filepath.Join(manifestDir, manifest.EntryPoint)
	if _, err := os.Stat(entryPath); err != nil {
		return nil, fmt.Errorf("plugin entry point not found: %s", entryPath)
	}

	return &manifest, nil
}
