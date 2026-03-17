// Package config provides configuration management
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"assistant/pkg/utils"

	"github.com/spf13/viper"
)

// Loader handles configuration loading
type Loader struct {
	configPath string
}

// NewLoader creates a new configuration loader
func NewLoader(configPath string) *Loader {
	return &Loader{
		configPath: configPath,
	}
}

// Load loads the configuration from file
func (l *Loader) Load() (*Config, error) {
	cfg := DefaultConfig()

	v := viper.New()
	v.SetConfigType("yaml")

	// Set default values
	setDefaults(v, cfg)

	// Try to find config file
	if l.configPath != "" {
		v.SetConfigFile(l.configPath)
	} else {
		// Look for config in standard locations
		v.SetConfigName("config")
		v.SetConfigType("yaml")

		// Check current directory
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")

		// Check home directory
		home, err := os.UserHomeDir()
		if err == nil {
			v.AddConfigPath(filepath.Join(home, ".assistant"))
		}

		// Check config directories
		v.AddConfigPath("/etc/assistant")
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults
	}

	// Unmarshal config
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Expand environment variables
	cfg.ExpandEnvVars()

	// Expand paths
	cfg.Storage.Path = utils.ExpandPath(cfg.Storage.Path)
	cfg.Plugins.Directory = utils.ExpandPath(cfg.Plugins.Directory)

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// setDefaults sets default values in viper
func setDefaults(v *viper.Viper, cfg *Config) {
	v.SetDefault("assistant.name", cfg.Assistant.Name)
	v.SetDefault("assistant.version", cfg.Assistant.Version)

	v.SetDefault("llm.provider", cfg.LLM.Provider)
	v.SetDefault("llm.model", cfg.LLM.Model)
	v.SetDefault("llm.api_key", cfg.LLM.APIKey)
	v.SetDefault("llm.base_url", cfg.LLM.BaseURL)
	v.SetDefault("llm.temperature", cfg.LLM.Temperature)
	v.SetDefault("llm.max_tokens", cfg.LLM.MaxTokens)

	v.SetDefault("plugins.enabled", cfg.Plugins.Enabled)
	v.SetDefault("plugins.directory", cfg.Plugins.Directory)

	v.SetDefault("mcp.enabled", cfg.MCP.Enabled)
	v.SetDefault("mcp.servers", cfg.MCP.Servers)

	v.SetDefault("permissions.default_mode", cfg.Permissions.DefaultMode)
	v.SetDefault("permissions.allowed_paths", cfg.Permissions.AllowedPaths)
	v.SetDefault("permissions.denied_commands", cfg.Permissions.DeniedCommands)

	v.SetDefault("storage.type", cfg.Storage.Type)
	v.SetDefault("storage.path", cfg.Storage.Path)

	v.SetDefault("logging.level", cfg.Logging.Level)
	v.SetDefault("logging.format", cfg.Logging.Format)
	v.SetDefault("logging.output", cfg.Logging.Output)
}

// LoadFromPath loads configuration from a specific path
func LoadFromPath(path string) (*Config, error) {
	loader := NewLoader(path)
	return loader.Load()
}

// LoadDefault loads configuration from default locations
func LoadDefault() (*Config, error) {
	loader := NewLoader("")
	return loader.Load()
}
