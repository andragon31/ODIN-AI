package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/odin-ai/odin/pkg/logger"
)

// Runtime manages plugin loading and execution
type Runtime struct {
	mu         sync.RWMutex
	pluginsDir string
	sandbox    bool
	instances  map[string]*PluginInstance
	config     *RuntimeConfig
}

// NewRuntime creates a new plugin runtime
func NewRuntime(cfg *RuntimeConfig) (*Runtime, error) {
	if cfg == nil {
		cfg = DefaultRuntimeConfig()
	}

	// Ensure plugins directory exists
	if err := os.MkdirAll(cfg.PluginsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugins directory: %w", err)
	}

	return &Runtime{
		pluginsDir: cfg.PluginsDir,
		sandbox:    cfg.Sandbox,
		instances:  make(map[string]*PluginInstance),
		config:     cfg,
	}, nil
}

// Load loads a plugin from the given path
func (r *Runtime) Load(pluginPath string) (Plugin, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate manifest
	manifest, err := ValidatePluginManifest(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin manifest: %w", err)
	}

	// Check if already loaded
	if instance, ok := r.instances[manifest.Name]; ok {
		logger.Debug("Plugin already loaded", "name", manifest.Name)
		return instance.plugin, nil
	}

	// Load the plugin based on type
	plugin, err := r.loadPlugin(manifest, filepath.Dir(pluginPath))
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %w", err)
	}

	// Store instance
	instance := &PluginInstance{
		Metadata: PluginMetadata{
			Name:        manifest.Name,
			Version:     manifest.Version,
			Author:      manifest.Author,
			Description: manifest.Description,
			Permissions: manifest.Permissions,
		},
		plugin:   plugin,
		runtime:  r,
		manifest: manifest,
	}

	r.instances[manifest.Name] = instance
	logger.Info("Plugin loaded", "name", manifest.Name, "version", manifest.Version)

	return plugin, nil
}

// loadPlugin loads the actual plugin implementation
func (r *Runtime) loadPlugin(manifest *Manifest, baseDir string) (Plugin, error) {
	entryPath := filepath.Join(baseDir, manifest.EntryPoint)

	// Check if WASM file
	if filepath.Ext(entryPath) == ".wasm" {
		return r.loadWASMPlugin(manifest, entryPath)
	}

	// For now, support dynamic library plugins (.so, .dll) via stub
	// Real implementation would use wasmtime or wasmer
	return r.loadStubPlugin(manifest, entryPath)
}

// loadWASMPlugin loads a WASM plugin using wasmtime
// This is a stub implementation - real implementation would use wasmtime-go
func (r *Runtime) loadWASMPlugin(manifest *Manifest, wasmPath string) (Plugin, error) {
	// Check if sandbox is enabled and validate permissions
	if r.sandbox {
		for _, perm := range manifest.Permissions {
			if !r.isAllowedPermission(perm) {
				return nil, fmt.Errorf("permission not allowed: %s", perm)
			}
		}
	}

	// Stub: Return a simple plugin that logs execution
	// Real implementation would use wasmtime to run the WASM module
	return &wasmPluginStub{
		manifest: manifest,
		wasmPath: wasmPath,
	}, nil
}

// loadStubPlugin provides a stub implementation for non-WASM plugins
func (r *Runtime) loadStubPlugin(manifest *Manifest, path string) (Plugin, error) {
	return &wasmPluginStub{
		manifest: manifest,
		wasmPath: path,
	}, nil
}

// wasmPluginStub is a stub implementation for WASM plugins
type wasmPluginStub struct {
	manifest *Manifest
	wasmPath string
}

// Metadata returns the plugin metadata
func (p *wasmPluginStub) Metadata() PluginMetadata {
	return PluginMetadata{
		Name:        p.manifest.Name,
		Version:     p.manifest.Version,
		Author:      p.manifest.Author,
		Description: p.manifest.Description,
		Permissions: p.manifest.Permissions,
	}
}

// Init initializes the plugin
func (p *wasmPluginStub) Init(ctx context.Context, config json.RawMessage) error {
	logger.Debug("Initializing stub plugin", "name", p.manifest.Name)
	// In real implementation, this would initialize the WASM module
	return nil
}

// Execute runs the plugin with the given input
func (p *wasmPluginStub) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	logger.Debug("Executing stub plugin", "name", p.manifest.Name)

	// In real implementation, this would call the WASM module's main function
	// For stub, just echo back the input
	output := map[string]interface{}{
		"plugin":    p.manifest.Name,
		"version":   p.manifest.Version,
		"executed":  true,
		"input":     string(input),
		"sandboxed": true,
	}

	return json.Marshal(output)
}

// Health checks if the plugin is healthy
func (p *wasmPluginStub) Health(ctx context.Context) error {
	// Check if WASM file still exists
	if _, err := os.Stat(p.wasmPath); err != nil {
		return fmt.Errorf("plugin file not accessible: %w", err)
	}
	return nil
}

// isAllowedPermission checks if a permission is allowed in sandbox mode
func (r *Runtime) isAllowedPermission(perm string) bool {
	allowedPermissions := []string{
		"fs:readonly",
		"fs:write",
		"net:allow",
		"net:disallow",
		"env:readonly",
		"env:write",
	}

	for _, allowed := range allowedPermissions {
		if allowed == perm {
			return true
		}
	}

	return false
}

// List returns all loaded plugin metadata
func (r *Runtime) List() ([]PluginMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata := make([]PluginMetadata, 0, len(r.instances))
	for _, instance := range r.instances {
		metadata = append(metadata, instance.Metadata)
	}

	return metadata, nil
}

// Execute runs a plugin by name with the given input
func (r *Runtime) Execute(ctx context.Context, name string, input json.RawMessage) (json.RawMessage, error) {
	r.mu.RLock()
	instance, ok := r.instances[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	return instance.plugin.Execute(ctx, input)
}

// Unload removes a plugin from the runtime
func (r *Runtime) Unload(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.instances[name]; !ok {
		return fmt.Errorf("plugin not loaded: %s", name)
	}

	delete(r.instances, name)
	logger.Info("Plugin unloaded", "name", name)

	return nil
}

// Install installs a plugin from a path
func (r *Runtime) Install(sourcePath string) error {
	// Validate manifest
	manifest, err := ValidatePluginManifest(sourcePath)
	if err != nil {
		return fmt.Errorf("invalid plugin manifest: %w", err)
	}

	// Determine destination
	destDir := filepath.Join(r.pluginsDir, manifest.Name)
	destManifest := filepath.Join(destDir, "manifest.json")

	// Check if already installed
	if _, err := os.Stat(destDir); err == nil && !r.config.Sandbox {
		// In non-sandbox mode, allow overwriting
		os.RemoveAll(destDir)
	}

	// Copy plugin files
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Copy manifest
	sourceData, _ := os.ReadFile(sourcePath)
	if err := os.WriteFile(destManifest, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to copy manifest: %w", err)
	}

	// Copy entry point
	sourceDir := filepath.Dir(sourcePath)
	entryDest := filepath.Join(destDir, manifest.EntryPoint)
	entrySrc := filepath.Join(sourceDir, manifest.EntryPoint)

	if err := copyFile(entrySrc, entryDest); err != nil {
		return fmt.Errorf("failed to copy entry point: %w", err)
	}

	logger.Info("Plugin installed", "name", manifest.Name, "version", manifest.Version)
	return nil
}

// Uninstall removes a plugin from the runtime and plugins directory
func (r *Runtime) Uninstall(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Unload if loaded
	if _, ok := r.instances[name]; ok {
		delete(r.instances, name)
	}

	// Remove from plugins directory
	pluginDir := filepath.Join(r.pluginsDir, name)
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("failed to remove plugin files: %w", err)
	}

	logger.Info("Plugin uninstalled", "name", name)
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
