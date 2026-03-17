// Package cli provides the CLI interface
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"assistant/internal/config"
	"assistant/internal/core"
	"assistant/internal/plugin"
	"assistant/internal/tool"

	"github.com/spf13/cobra"
)

// App represents the CLI application
type App struct {
	rootCmd *cobra.Command
	engine  *core.Engine
	config  *config.Config
}

// NewApp creates a new CLI application
func NewApp() *App {
	app := &App{
		rootCmd: &cobra.Command{
			Use:   "assistant",
			Short: "AI Assistant CLI",
			Long:  "An AI-powered assistant for daily tasks, code writing, and automation",
		},
	}

	// Add subcommands
	app.rootCmd.AddCommand(
		app.chatCmd(),
		app.configCmd(),
		app.pluginCmd(),
		app.mcpCmd(),
		app.versionCmd(),
	)

	// Set version
	app.rootCmd.Version = "1.0.0"

	return app
}

// Run executes the CLI application
func (a *App) Run() error {
	return a.rootCmd.Execute()
}

// chatCmd returns the chat command
func (a *App) chatCmd() *cobra.Command {
	var sessionID string
	var configFile string

	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Start an interactive chat session",
		Long:  "Start an interactive chat session with the AI assistant",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			var err error
			if configFile != "" {
				a.config, err = config.LoadFromPath(configFile)
			} else {
				a.config, err = config.LoadDefault()
			}
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create engine
			a.engine = core.NewEngine(a.config)

			// Register built-in tools
			a.registerBuiltinTools()

			// Initialize engine
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if err := a.engine.Initialize(ctx); err != nil {
				return fmt.Errorf("failed to initialize engine: %w", err)
			}
			defer a.engine.Close()

			// Start REPL
			repl := NewREPL(a.engine, a.config)
			repl.SessionID = sessionID

			// Handle signals
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigCh
				cancel()
				repl.Close()
			}()

			return repl.Run(ctx)
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "Session ID to continue")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Config file path")

	return cmd
}

// configCmd returns the config command
func (a *App) configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "show",
			Short: "Show current configuration",
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := config.LoadDefault()
				if err != nil {
					return err
				}
				printConfig(cfg)
				return nil
			},
		},
		&cobra.Command{
			Use:   "init",
			Short: "Initialize configuration file",
			RunE: func(cmd *cobra.Command, args []string) error {
				return initConfig()
			},
		},
	)

	return cmd
}

// pluginCmd returns the plugin command
func (a *App) pluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage plugins",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List available plugins",
			RunE: func(cmd *cobra.Command, args []string) error {
				factories := plugin.ListFactories()
				fmt.Println("Available plugins:")
				for _, name := range factories {
					fmt.Printf("  - %s\n", name)
				}
				return nil
			},
		},
	)

	return cmd
}

// mcpCmd returns the MCP command
func (a *App) mcpCmd() *cobra.Command {
	var addr string

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server commands",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "serve",
			Short: "Start MCP server",
			RunE: func(cmd *cobra.Command, args []string) error {
				return startMCPServer(addr)
			},
		},
	)

	cmd.PersistentFlags().StringVarP(&addr, "addr", "a", ":8080", "Server address")

	return cmd
}

// versionCmd returns the version command
func (a *App) versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("AI Assistant v1.0.0")
		},
	}
}

// registerBuiltinTools registers built-in tools
func (a *App) registerBuiltinTools() {
	// File tools
	plugin.RegisterFactory("file", func() plugin.Plugin {
		return &builtinPlugin{
			info: plugin.PluginInfo{
				Name:        "file",
				Version:     "1.0.0",
				Description: "File operations",
			},
			tools: []plugin.Tool{
				tool.NewFileReadTool(),
				tool.NewFileWriteTool(),
				tool.NewFileDeleteTool(),
				tool.NewFileListTool(),
				tool.NewFileExistsTool(),
			},
		}
	})

	// Shell tools
	plugin.RegisterFactory("shell", func() plugin.Plugin {
		return &builtinPlugin{
			info: plugin.PluginInfo{
				Name:        "shell",
				Version:     "1.0.0",
				Description: "Shell command execution",
			},
			tools: []plugin.Tool{
				tool.NewShellExecuteTool(),
				tool.NewShellBackgroundTool(),
				tool.NewShellKillTool(),
			},
		}
	})

	// Code tools
	plugin.RegisterFactory("code", func() plugin.Plugin {
		return &builtinPlugin{
			info: plugin.PluginInfo{
				Name:        "code",
				Version:     "1.0.0",
				Description: "Code execution and tools",
			},
			tools: []plugin.Tool{
				tool.NewCodeRunTool(),
				tool.NewCodeFormatTool(),
				tool.NewCodeLintTool(),
				tool.NewCodeTestTool(),
			},
		}
	})

	// Git tools
	plugin.RegisterFactory("git", func() plugin.Plugin {
		return &builtinPlugin{
			info: plugin.PluginInfo{
				Name:        "git",
				Version:     "1.0.0",
				Description: "Git operations",
			},
			tools: []plugin.Tool{
				tool.NewGitStatusTool(),
				tool.NewGitLogTool(),
				tool.NewGitDiffTool(),
				tool.NewGitCommitTool(),
				tool.NewGitBranchTool(),
			},
		}
	})

	// Web tools
	plugin.RegisterFactory("web", func() plugin.Plugin {
		return &builtinPlugin{
			info: plugin.PluginInfo{
				Name:        "web",
				Version:     "1.0.0",
				Description: "Web requests and API calls",
			},
			tools: []plugin.Tool{
				tool.NewWebFetchTool(),
				tool.NewWebSearchTool(),
				tool.NewWebAPITool(),
			},
		}
	})

	// Email tools
	plugin.RegisterFactory("email", func() plugin.Plugin {
		return &builtinPlugin{
			info: plugin.PluginInfo{
				Name:        "email",
				Version:     "1.0.0",
				Description: "Email operations",
			},
			tools: []plugin.Tool{
				tool.NewEmailSendTool(tool.EmailConfig{}),
				tool.NewEmailReadTool(tool.EmailConfig{}),
				tool.NewEmailListTool(tool.EmailConfig{}),
			},
		}
	})
}

// builtinPlugin implements Plugin for built-in tools
type builtinPlugin struct {
	info  plugin.PluginInfo
	tools []plugin.Tool
}

func (p *builtinPlugin) Info() plugin.PluginInfo {
	return p.info
}

func (p *builtinPlugin) Tools() []plugin.Tool {
	return p.tools
}

func (p *builtinPlugin) Init(ctx context.Context, config map[string]interface{}) error {
	return nil
}

func (p *builtinPlugin) Close() error {
	return nil
}
