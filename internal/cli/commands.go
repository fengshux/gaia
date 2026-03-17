// Package cli provides the CLI interface
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"assistant/internal/config"
	"assistant/internal/mcp"
	"assistant/internal/tool"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// printConfig prints the configuration
func printConfig(cfg *config.Config) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Bold(true)

	fmt.Println(style.Render("Configuration:"))
	fmt.Printf("  Name: %s\n", cfg.Assistant.Name)
	fmt.Printf("  Version: %s\n", cfg.Assistant.Version)
	fmt.Println()
	fmt.Println(style.Render("LLM:"))
	fmt.Printf("  Provider: %s\n", cfg.LLM.Provider)
	fmt.Printf("  Model: %s\n", cfg.LLM.Model)
	fmt.Printf("  Base URL: %s\n", cfg.LLM.BaseURL)
	fmt.Printf("  Temperature: %.2f\n", cfg.LLM.Temperature)
	fmt.Printf("  Max Tokens: %d\n", cfg.LLM.MaxTokens)
	fmt.Println()
	fmt.Println(style.Render("Plugins:"))
	fmt.Printf("  Enabled: %v\n", cfg.Plugins.Enabled)
	fmt.Printf("  Directory: %s\n", cfg.Plugins.Directory)
	fmt.Println()
	fmt.Println(style.Render("Storage:"))
	fmt.Printf("  Type: %s\n", cfg.Storage.Type)
	fmt.Printf("  Path: %s\n", cfg.Storage.Path)
	fmt.Println()
	fmt.Println(style.Render("Permissions:"))
	fmt.Printf("  Default Mode: %s\n", cfg.Permissions.DefaultMode)
	fmt.Printf("  Allowed Paths: %v\n", cfg.Permissions.AllowedPaths)
	fmt.Printf("  Denied Commands: %v\n", cfg.Permissions.DeniedCommands)
}

// initConfig initializes a configuration file
func initConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".assistant")
	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists: %s", configPath)
	}

	// Create directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create default config
	cfg := config.DefaultConfig()
	cfg.Storage.Path = filepath.Join(configDir, "assistant.db")
	cfg.Plugins.Directory = filepath.Join(configDir, "plugins")

	// Write config
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Created config file: %s\n", configPath)
	fmt.Println("\nPlease edit the config file and set your API key:")
	fmt.Printf("  export OPENAI_API_KEY=your-api-key\n")
	fmt.Printf("  vim %s\n", configPath)

	return nil
}

// startMCPServer starts the MCP server
func startMCPServer(addr string) error {
	server := mcp.NewServer("AI Assistant", "1.0.0")

	// Register tools
	server.RegisterTool(mcp.ToolDefinition{
		Name:        "file_read",
		Description: "Read the contents of a file",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"path": {Type: "string", Description: "The path to the file"},
			},
		},
	}, func(ctx context.Context, params map[string]interface{}) (*mcp.ToolCallResult, error) {
		t := tool.NewFileReadTool()
		result, err := t.Execute(ctx, params)
		if err != nil {
			return &mcp.ToolCallResult{
				Content: []mcp.Content{{Type: "text", Text: err.Error()}},
				IsError: true,
			}, nil
		}
		return &mcp.ToolCallResult{
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("%v", result)}},
		}, nil
	})

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "file_write",
		Description: "Write content to a file",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"path":    {Type: "string", Description: "The path to the file"},
				"content": {Type: "string", Description: "The content to write"},
			},
		},
	}, func(ctx context.Context, params map[string]interface{}) (*mcp.ToolCallResult, error) {
		t := tool.NewFileWriteTool()
		result, err := t.Execute(ctx, params)
		if err != nil {
			return &mcp.ToolCallResult{
				Content: []mcp.Content{{Type: "text", Text: err.Error()}},
				IsError: true,
			}, nil
		}
		return &mcp.ToolCallResult{
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("%v", result)}},
		}, nil
	})

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "shell_execute",
		Description: "Execute a shell command",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"command": {Type: "string", Description: "The command to execute"},
			},
		},
	}, func(ctx context.Context, params map[string]interface{}) (*mcp.ToolCallResult, error) {
		t := tool.NewShellExecuteTool()
		result, err := t.Execute(ctx, params)
		if err != nil {
			return &mcp.ToolCallResult{
				Content: []mcp.Content{{Type: "text", Text: err.Error()}},
				IsError: true,
			}, nil
		}
		return &mcp.ToolCallResult{
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("%v", result)}},
		}, nil
	})

	fmt.Printf("Starting MCP server on %s\n", addr)
	fmt.Println("Endpoints:")
	fmt.Printf("  WebSocket: ws://localhost%s/mcp\n", addr)
	fmt.Printf("  Health:    http://localhost%s/health\n", addr)

	return server.StartHTTP(addr)
}
