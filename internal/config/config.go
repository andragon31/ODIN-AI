// Package config provides configuration management for ODIN
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the ODIN configuration
type Config struct {
	Version       string         `mapstructure:"version"`
	Mode          string         `mapstructure:"mode"`
	HomeDir       string         `mapstructure:"home_dir"`
	Memory        MemoryConfig   `mapstructure:"memory"`
	Sync          SyncConfig     `mapstructure:"sync"`
	Guardian      GuardianConfig `mapstructure:"guardian"`
	Router        RouterConfig   `mapstructure:"router"`
	Discovery     DiscoveryConfig `mapstructure:"discovery"`
	Observability ObservConfig   `mapstructure:"observability"`
	Plugins       PluginsConfig  `mapstructure:"plugins"`
	Themes        ThemesConfig   `mapstructure:"themes"`
	Session       SessionConfig  `mapstructure:"session"`
	Runes         RunesConfig    `mapstructure:"runes"`
	Verify        VerifyConfig   `mapstructure:"verify"`
}

// DiscoveryConfig holds the results of tool discovery
type DiscoveryConfig struct {
	LastScan    string                     `mapstructure:"last_scan"`
	Tools       map[string]DiscoveryResult `mapstructure:"tools"`
}

// ToolModel represents a detected model from an external tool
type ToolModel struct {
	Name        string `json:"name" mapstructure:"name"`
	Provider    string `json:"provider" mapstructure:"provider"`
	DisplayName string `json:"displayName" mapstructure:"displayName"`
}

// DiscoveryResult represents the detected configuration for a specific tool
type DiscoveryResult struct {
	ToolName string            `json:"toolName" mapstructure:"toolName"`
	Models   []ToolModel       `json:"models" mapstructure:"models"`
	APIKeys  map[string]string `json:"apiKeys" mapstructure:"apiKeys"`
	Path     string            `json:"path" mapstructure:"path"`
}

// MemoryConfig holds memory engine configuration
type MemoryConfig struct {
	Engine     string        `mapstructure:"engine"`
	Path       string        `mapstructure:"path"`
	Encryption bool          `mapstructure:"encryption"`
	Pruning    PruningConfig `mapstructure:"pruning"`
}

// PruningConfig holds pruning configuration
type PruningConfig struct {
	KeepTags []string `mapstructure:"keep_tags"`
	Interval string   `mapstructure:"interval"`
}

// SyncConfig holds sync engine configuration
type SyncConfig struct {
	Backend  string `mapstructure:"backend"`
	Remote   string `mapstructure:"remote"`
	AutoPush bool   `mapstructure:"auto_push"`
	GPGSign  bool   `mapstructure:"gpg_sign"`
}

// GuardianConfig holds security guardian configuration
type GuardianConfig struct {
	PolicyEngine string     `mapstructure:"policy_engine"`
	RulesPath    string     `mapstructure:"rules_path"`
	SAST         SASTConfig `mapstructure:"saast"`
	BlockOnCrit  bool       `mapstructure:"block_on_critical"`
}

// SASTConfig holds SAST tool configuration
type SASTConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Tools   []string `mapstructure:"tools"`
}

// RouterConfig holds model router configuration
type RouterConfig struct {
	Default    string                    `mapstructure:"default"`
	Fallback   []string                  `mapstructure:"fallback"`
	CostCapDay float64                   `mapstructure:"cost_cap_daily"`
	Mapping    ModelMapping              `mapstructure:"mapping"`
	Providers  map[string]ProviderConfig `mapstructure:"providers"`
}

// ProviderConfig holds configuration for a provider
type ProviderConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
	APIKey   string `mapstructure:"api_key"`
}

// OllamaConfig holds Ollama-specific configuration
type OllamaConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

// OpenRouterConfig holds OpenRouter-specific configuration
type OpenRouterConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	APIKey   string `mapstructure:"api_key"`
	Endpoint string `mapstructure:"endpoint"`
}

// AnthropicConfig holds Anthropic-specific configuration
type AnthropicConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	APIKey   string `mapstructure:"api_key"`
	Endpoint string `mapstructure:"endpoint"`
}

// OpenAIConfig holds OpenAI-specific configuration
type OpenAIConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	APIKey   string `mapstructure:"api_key"`
	Endpoint string `mapstructure:"endpoint"`
}

// Default endpoints
const (
	DefaultOllamaEndpoint     = "http://localhost:11434"
	DefaultOpenRouterEndpoint = "https://openrouter.ai/api/v1"
	DefaultAnthropicEndpoint  = "https://api.anthropic.com"
	DefaultOpenAIEndpoint     = "https://api.openai.com/v1"
)

// ModelMapping holds manual mappings for agents/models
type ModelMapping struct {
	PhaseMappings map[string]string `mapstructure:"phase_mappings"`
	RuneMappings  map[string]string `mapstructure:"rune_mappings"`
}

// SDD Phases for mapping
var SDDPhases = []string{
	"explore",
	"propose",
	"spec",
	"design",
	"tasks",
	"apply",
	"verify",
	"archive",
}

// Common Runes for mapping
var CommonRunes = []string{
	"branch-pr",
	"issue-creation",
	"judgment-day",
	"skill-creator",
}

// ObservConfig holds observability configuration
type ObservConfig struct {
	MetricsPort int    `mapstructure:"metrics_port"`
	LogLevel    string `mapstructure:"log_level"`
	LogPath     string `mapstructure:"log_path"`
}

// PluginsConfig holds plugins configuration
type PluginsConfig struct {
	Sandbox    bool     `mapstructure:"sandbox"`
	AutoUpdate bool     `mapstructure:"auto_update"`
	Allowed    []string `mapstructure:"allowed_paths"`
}

// ThemesConfig holds themes configuration
type ThemesConfig struct {
	Current string `mapstructure:"current"`
	Path    string `mapstructure:"path"`
}

// SessionConfig holds session configuration
type SessionConfig struct {
	Path             string `mapstructure:"path"`
	SnapshotInterval string `mapstructure:"snapshot_interval"`
	MaxSessions      int    `mapstructure:"max_sessions"`
}

// RunesConfig holds Runes skills registry configuration
type RunesConfig struct {
	CachePath  string `mapstructure:"cache_path"`
	Sandbox    bool   `mapstructure:"sandbox"`
	AutoUpdate bool   `mapstructure:"auto_update"`
}

// VerifyConfig holds Nornir verification configuration
type VerifyConfig struct {
	Timeout        string   `mapstructure:"timeout"`
	FlakyThreshold int      `mapstructure:"flaky_threshold"`
	MatrixTargets  []string `mapstructure:"matrix_targets"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		Version: "1.0",
		Mode:    "local",
		HomeDir: homeDir,
		Memory: MemoryConfig{
			Engine:     "sqlite-vss",
			Path:       filepath.Join(homeDir, ".odin", "memory.db"),
			Encryption: true,
			Pruning: PruningConfig{
				KeepTags: []string{"arch", "spec", "security"},
				Interval: "24h",
			},
		},
		Sync: SyncConfig{
			Backend:  "git",
			AutoPush: false,
			GPGSign:  false,
		},
		Guardian: GuardianConfig{
			PolicyEngine: "opa",
			RulesPath:    filepath.Join(homeDir, ".odin", "rules"),
			SAST: SASTConfig{
				Enabled: true,
				Tools:   []string{"gosec", "semgrep"},
			},
			BlockOnCrit: true,
		},
		Router: RouterConfig{
			Default:    "ollama-local",
			Fallback:   []string{"openrouter", "anthropic"},
			CostCapDay: 0.0,
			Mapping: ModelMapping{
				PhaseMappings: make(map[string]string),
				RuneMappings:  make(map[string]string),
			},
			Providers: map[string]ProviderConfig{
				"ollama": {
					Enabled:  true,
					Endpoint: DefaultOllamaEndpoint,
				},
				"openrouter": {
					Enabled:  true,
					Endpoint: DefaultOpenRouterEndpoint,
				},
				"anthropic": {
					Enabled:  true,
					Endpoint: DefaultAnthropicEndpoint,
				},
				"openai": {
					Enabled:  false,
					Endpoint: DefaultOpenAIEndpoint,
				},
			},
		},
		Discovery: DiscoveryConfig{
			Tools: make(map[string]DiscoveryResult),
		},
		Observability: ObservConfig{
			MetricsPort: 9090,
			LogLevel:    "info",
			LogPath:     filepath.Join(homeDir, ".odin", "logs"),
		},
		Plugins: PluginsConfig{
			Sandbox:    true,
			AutoUpdate: false,
			Allowed:    []string{filepath.Join(homeDir, ".odin", "plugins")},
		},
		Themes: ThemesConfig{
			Current: "rose-pine",
			Path:    filepath.Join(homeDir, ".odin", "themes"),
		},
		Session: SessionConfig{
			Path:             filepath.Join(homeDir, ".odin", "sessions"),
			SnapshotInterval: "5m",
			MaxSessions:      10,
		},
		Runes: RunesConfig{
			CachePath:  filepath.Join(homeDir, ".odin", "runes"),
			Sandbox:    true,
			AutoUpdate: false,
		},
		Verify: VerifyConfig{
			Timeout:        "15m",
			FlakyThreshold: 3,
			MatrixTargets:  []string{"ubuntu", "macos", "windows"},
		},
	}
}

// Load loads configuration from file and environment
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	defaults := DefaultConfig()
	v.SetDefault("version", defaults.Version)
	v.SetDefault("mode", defaults.Mode)
	v.SetDefault("memory", defaults.Memory)
	v.SetDefault("sync", defaults.Sync)
	v.SetDefault("guardian", defaults.Guardian)
	v.SetDefault("router", defaults.Router)
	v.SetDefault("discovery", defaults.Discovery)
	v.SetDefault("observability", defaults.Observability)
	v.SetDefault("plugins", defaults.Plugins)
	v.SetDefault("themes", defaults.Themes)
	v.SetDefault("session", defaults.Session)
	v.SetDefault("runes", defaults.Runes)
	v.SetDefault("verify", defaults.Verify)

	// Support environment variables
	v.SetEnvPrefix("ODIN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Load config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// EnsureDirs creates necessary directories
func (c *Config) EnsureDirs() error {
	// Calculate sync directory path
	syncRepoPath := filepath.Join(c.HomeDir, ".odin", "config")

	dirs := []string{
		filepath.Dir(c.Memory.Path),
		c.Guardian.RulesPath,
		c.Observability.LogPath,
		c.Plugins.Allowed[0],
		c.Themes.Path,
		c.Session.Path,
		c.Runes.CachePath,
		syncRepoPath,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// Save persists the current configuration to the specified path
func (c *Config) Save(configPath string) error {
	v := viper.New()

	// Set all values into viper
	v.Set("version", c.Version)
	v.Set("mode", c.Mode)
	v.Set("home_dir", c.HomeDir)
	v.Set("memory", c.Memory)
	v.Set("sync", c.Sync)
	v.Set("guardian", c.Guardian)
	v.Set("router", c.Router)
	v.Set("discovery", c.Discovery)
	v.Set("observability", c.Observability)
	v.Set("plugins", c.Plugins)
	v.Set("themes", c.Themes)
	v.Set("session", c.Session)
	v.Set("runes", c.Runes)
	v.Set("verify", c.Verify)

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
