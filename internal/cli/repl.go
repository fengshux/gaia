// Package cli provides the CLI interface
package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"partner/internal/config"
	"partner/internal/core"

	"github.com/charmbracelet/lipgloss"
)

// REPL provides an interactive read-eval-print loop
type REPL struct {
	engine    *core.Engine
	config    *config.Config
	reader    *bufio.Reader
	SessionID string
	running   bool

	// Styles
	promptStyle lipgloss.Style
	userStyle   lipgloss.Style
	aiStyle     lipgloss.Style
	errorStyle  lipgloss.Style
	toolStyle   lipgloss.Style
}

// NewREPL creates a new REPL
func NewREPL(engine *core.Engine, config *config.Config) *REPL {
	return &REPL{
		engine: engine,
		config: config,
		reader: bufio.NewReader(os.Stdin),
		promptStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true),
		userStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")),
		aiStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")),
		toolStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")),
	}
}

// Run starts the REPL
func (r *REPL) Run(ctx context.Context) error {
	r.running = true

	// Print welcome message
	r.printWelcome()

	for r.running {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Read input
			input, err := r.readLine()
			if err != nil {
				return err
			}

			// Handle commands
			if strings.HasPrefix(input, "/") {
				if err := r.handleCommand(ctx, input); err != nil {
					r.printError(err.Error())
				}
				continue
			}

			// Skip empty input
			if strings.TrimSpace(input) == "" {
				continue
			}

			// Process input
			if err := r.processInput(ctx, input); err != nil {
				r.printError(err.Error())
			}
		}
	}

	return nil
}

// Close closes the REPL
func (r *REPL) Close() {
	r.running = false
}

// readLine reads a line of input
func (r *REPL) readLine() (string, error) {
	fmt.Print(r.promptStyle.Render(">>> "))
	input, err := r.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// handleCommand handles slash commands
func (r *REPL) handleCommand(ctx context.Context, input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "/help", "/h", "/?":
		r.printHelp()
	case "/exit", "/quit", "/q":
		r.running = false
		fmt.Println("Goodbye!")
	case "/clear":
		r.clearSession()
	case "/session":
		r.handleSessionCommand(args)
	case "/tools":
		r.listTools()
	case "/config":
		r.showConfig()
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}

	return nil
}

// processInput processes user input
func (r *REPL) processInput(ctx context.Context, input string) error {
	// Print user message
	fmt.Println(r.userStyle.Render("You: ") + input)

	// Process with engine
	response, err := r.engine.ProcessWithSession(ctx, input, r.SessionID)
	if err != nil {
		return err
	}

	// Update session ID
	session := r.engine.GetSessionManager().GetCurrent()
	if session != nil {
		r.SessionID = session.ID
	}

	// Print response
	if response.Content != "" {
		fmt.Println(r.aiStyle.Render("Assistant: ") + response.Content)
	}

	// Print tool calls if any
	if len(response.ToolCalls) > 0 {
		for _, tc := range response.ToolCalls {
			fmt.Println(r.toolStyle.Render(fmt.Sprintf("[Tool: %s]", tc.Name)))
		}
	}

	return nil
}

// printWelcome prints the welcome message
func (r *REPL) printWelcome() {
	welcome := `
╔═══════════════════════════════════════════╗
║         AI Assistant v1.0.0               ║
║                                           ║
║  Type /help for available commands        ║
║  Press Ctrl+C to exit                     ║
╚═══════════════════════════════════════════╝
`
	fmt.Println(r.aiStyle.Render(welcome))
}

// printHelp prints help information
func (r *REPL) printHelp() {
	help := `
Available Commands:
  /help, /h, /?    Show this help message
  /exit, /quit, /q Exit the assistant
  /clear           Clear current session
  /session         Manage sessions (list, new, switch)
  /tools           List available tools
  /config          Show current configuration

Examples:
  Read a file:     Read the contents of go.mod
  Run a command:   Execute 'ls -la'
  Ask a question:  What is the capital of France?
`
	fmt.Println(help)
}

// printError prints an error message
func (r *REPL) printError(msg string) {
	fmt.Println(r.errorStyle.Render("Error: ") + msg)
}

// clearSession clears the current session
func (r *REPL) clearSession() {
	r.SessionID = ""
	r.engine.GetSessionManager().Create("New Session")
	fmt.Println("Session cleared.")
}

// handleSessionCommand handles session commands
func (r *REPL) handleSessionCommand(args []string) {
	if len(args) == 0 {
		// Show current session
		session := r.engine.GetSessionManager().GetCurrent()
		if session != nil {
			fmt.Printf("Current session: %s (%s)\n", session.Name, session.ID)
		} else {
			fmt.Println("No active session")
		}
		return
	}

	switch args[0] {
	case "list":
		sessions := r.engine.GetSessionManager().List()
		fmt.Println("Sessions:")
		for _, s := range sessions {
			current := ""
			if s.ID == r.SessionID {
				current = " (current)"
			}
			fmt.Printf("  %s - %s%s\n", s.ID, s.Name, current)
		}
	case "new":
		name := "New Session"
		if len(args) > 1 {
			name = strings.Join(args[1:], " ")
		}
		session := r.engine.GetSessionManager().Create(name)
		r.SessionID = session.ID
		fmt.Printf("Created session: %s\n", session.ID)
	case "switch":
		if len(args) < 2 {
			fmt.Println("Usage: /session switch <id>")
			return
		}
		if r.engine.GetSessionManager().SetCurrent(args[1]) {
			r.SessionID = args[1]
			fmt.Printf("Switched to session: %s\n", args[1])
		} else {
			fmt.Println("Session not found")
		}
	default:
		fmt.Println("Usage: /session [list|new|switch]")
	}
}

// listTools lists available tools
func (r *REPL) listTools() {
	tools := r.engine.GetPluginManager().GetAllTools()
	fmt.Println("Available Tools:")
	for _, t := range tools {
		fmt.Printf("  %-20s %s\n", t.Name, t.Description)
	}
}

// showConfig shows current configuration
func (r *REPL) showConfig() {
	fmt.Printf("Model: %s\n", r.config.LLM.Model)
	fmt.Printf("Provider: %s\n", r.config.LLM.Provider)
	fmt.Printf("Base URL: %s\n", r.config.LLM.BaseURL)
	fmt.Printf("Temperature: %.2f\n", r.config.LLM.Temperature)
	fmt.Printf("Max Tokens: %d\n", r.config.LLM.MaxTokens)
	fmt.Printf("Plugins: %v\n", r.config.Plugins.Enabled)
}

// ProcessStream processes input with streaming output
func (r *REPL) ProcessStream(ctx context.Context, input string) error {
	stream, err := r.engine.ProcessStreamWithSession(ctx, input, r.SessionID)
	if err != nil {
		return err
	}

	fmt.Print(r.aiStyle.Render("Assistant: "))

	for chunk := range stream {
		if chunk.Error != nil {
			r.printError(chunk.Error.Error())
			break
		}

		fmt.Print(chunk.Content)

		if chunk.Done {
			fmt.Println()
			break
		}
	}

	// Update session ID
	session := r.engine.GetSessionManager().GetCurrent()
	if session != nil {
		r.SessionID = session.ID
	}

	return nil
}
