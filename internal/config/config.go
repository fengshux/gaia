// Package config provides configuration management
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config represents the main configuration
type Config struct {
	Gaia       GaiaConfig       `mapstructure:"gaia"`
	LLM        LLMConfig        `mapstructure:"llm"`
	Plugins    PluginsConfig    `mapstructure:"plugins"`
	MCP        MCPConfig        `mapstructure:"mcp"`
	Permissions PermissionsConfig `mapstructure:"permissions"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Logging    LoggingConfig    `mapstructure:"logging"`
}

// GaiaConfig represents Gaia configuration
type GaiaConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

// LLMConfig represents LLM provider configuration
type LLMConfig struct {
	Provider    string  `mapstructure:"provider"`
	Model       string  `mapstructure:"model"`
	APIKey      string  `mapstructure:"api_key"`
	BaseURL     string  `mapstructure:"base_url"`
	Temperature float64 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`
}

// PluginsConfig represents plugins configuration
type PluginsConfig struct {
	Enabled   []string `mapstructure:"enabled"`
	Directory string   `mapstructure:"directory"`
}

// MCPConfig represents MCP configuration
type MCPConfig struct {
	Enabled bool             `mapstructure:"enabled"`
	Servers []MCPServerConfig `mapstructure:"servers"`
}

// MCPServerConfig represents MCP server configuration
type MCPServerConfig struct {
	Name    string   `mapstructure:"name"`
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
}

// PermissionsConfig represents permissions configuration
type PermissionsConfig struct {
	DefaultMode      string   `mapstructure:"default_mode"`
	AllowedPaths     []string `mapstructure:"allowed_paths"`
	DeniedCommands   []string `mapstructure:"denied_commands"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Gaia: GaiaConfig{
			Name:    "Gaia",
			Version: "1.0.0",
		},
		LLM: LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			APIKey:      "${OPENAI_API_KEY}",
			BaseURL:     "https://api.openai.com/v1",
			Temperature: 0.7,
			MaxTokens:   4096,
		},
		Plugins: PluginsConfig{
			Enabled:   []string{"file", "shell", "code"},
			Directory: "./plugins",
		},
		MCP: MCPConfig{
			Enabled: true,
			Servers: []MCPServerConfig{},
		},
		Permissions: PermissionsConfig{
			DefaultMode:    "ask",
			AllowedPaths:   []string{},
			DeniedCommands: []string{"rm -rf", "sudo"},
		},
		Storage: StorageConfig{
			Type: "sqlite",
			Path: "./data/gaia.db",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
	}
}

// ExpandEnvVars expands environment variables in the configuration
func (c *Config) ExpandEnvVars() {
	c.LLM.APIKey = expandEnv(c.LLM.APIKey)
	c.LLM.BaseURL = expandEnv(c.LLM.BaseURL)
	c.Storage.Path = expandEnv(c.Storage.Path)
	c.Plugins.Directory = expandEnv(c.Plugins.Directory)
}

// expandEnv expands environment variables in a string
func expandEnv(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		envVar := s[2 : len(s)-1]
		return os.Getenv(envVar)
	}
	return os.ExpandEnv(s)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.LLM.APIKey == "" {
		return fmt.Errorf("LLM API key is required")
	}
	if c.LLM.Model == "" {
		return fmt.Errorf("LLM model is required")
	}
	if c.Storage.Path == "" {
		return fmt.Errorf("storage path is required")
	}
	return nil
}
