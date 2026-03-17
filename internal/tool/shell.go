// Package tool provides built-in tools
package tool

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ShellExecuteTool executes shell commands
type ShellExecuteTool struct {
	timeout time.Duration
}

// NewShellExecuteTool creates a new shell execute tool
func NewShellExecuteTool() *ShellExecuteTool {
	return &ShellExecuteTool{
		timeout: 60 * time.Second,
	}
}

// WithTimeout sets the command timeout
func (t *ShellExecuteTool) WithTimeout(timeout time.Duration) *ShellExecuteTool {
	t.timeout = timeout
	return t
}

func (t *ShellExecuteTool) Name() string {
	return "shell_execute"
}

func (t *ShellExecuteTool) Description() string {
	return "Execute a shell command"
}

func (t *ShellExecuteTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"command": StringParam{
				Type:        "string",
				Description: "The command to execute",
			},
			"cwd": StringParam{
				Type:        "string",
				Description: "Working directory for the command (optional)",
			},
			"timeout": NumberParam{
				Type:        "number",
				Description: "Timeout in seconds (default: 60)",
			},
			"env": ObjectParam{
				Type:        "object",
				Description: "Environment variables as key-value pairs",
				Properties:  map[string]interface{}{},
			},
		},
		Required: []string{"command"},
	}
}

func (t *ShellExecuteTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	command, ok := params["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command parameter is required")
	}

	// Determine shell based on OS
	shell := "/bin/sh"
	shellFlag := "-c"
	if _, err := os.Stat("/bin/bash"); err == nil {
		shell = "/bin/bash"
	}

	// Create command
	cmd := exec.CommandContext(ctx, shell, shellFlag, command)

	// Set working directory
	if cwd, ok := params["cwd"].(string); ok {
		cmd.Dir = cwd
	}

	// Set environment variables
	if env, ok := params["env"].(map[string]interface{}); ok {
		for k, v := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%v", k, v))
		}
	}

	// Set timeout
	timeout := t.timeout
	if t, ok := params["timeout"].(float64); ok {
		timeout = time.Duration(t) * time.Second
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := map[string]interface{}{
		"command":  command,
		"stdout":   stdout.String(),
		"stderr":   stderr.String(),
		"duration": duration.String(),
	}

	if err != nil {
		result["error"] = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			result["exitCode"] = exitErr.ExitCode()
		}
	} else {
		result["exitCode"] = 0
	}

	return result, nil
}

// ShellBackgroundTool runs a command in the background
type ShellBackgroundTool struct{}

// NewShellBackgroundTool creates a new background shell tool
func NewShellBackgroundTool() *ShellBackgroundTool {
	return &ShellBackgroundTool{}
}

func (t *ShellBackgroundTool) Name() string {
	return "shell_background"
}

func (t *ShellBackgroundTool) Description() string {
	return "Run a shell command in the background"
}

func (t *ShellBackgroundTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"command": StringParam{
				Type:        "string",
				Description: "The command to run in the background",
			},
			"cwd": StringParam{
				Type:        "string",
				Description: "Working directory for the command (optional)",
			},
		},
		Required: []string{"command"},
	}
}

func (t *ShellBackgroundTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	command, ok := params["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command parameter is required")
	}

	// Determine shell
	shell := "/bin/sh"
	shellFlag := "-c"
	if _, err := os.Stat("/bin/bash"); err == nil {
		shell = "/bin/bash"
	}

	// Create command
	cmd := exec.Command(shell, shellFlag, command)

	// Set working directory
	if cwd, ok := params["cwd"].(string); ok {
		cmd.Dir = cwd
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	return map[string]interface{}{
		"pid":     cmd.Process.Pid,
		"command": command,
		"status":  "running",
	}, nil
}

// ShellKillTool kills a process
type ShellKillTool struct{}

// NewShellKillTool creates a new kill tool
func NewShellKillTool() *ShellKillTool {
	return &ShellKillTool{}
}

func (t *ShellKillTool) Name() string {
	return "shell_kill"
}

func (t *ShellKillTool) Description() string {
	return "Kill a process by PID"
}

func (t *ShellKillTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"pid": NumberParam{
				Type:        "number",
				Description: "Process ID to kill",
			},
			"signal": StringParam{
				Type:        "string",
				Description: "Signal to send (default: SIGTERM)",
				Enum:        []string{"SIGTERM", "SIGKILL", "SIGINT"},
			},
		},
		Required: []string{"pid"},
	}
}

func (t *ShellKillTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	pid, ok := params["pid"].(float64)
	if !ok {
		return nil, fmt.Errorf("pid parameter is required")
	}

	signal := "SIGTERM"
	if s, ok := params["signal"].(string); ok {
		signal = strings.ToUpper(s)
	}

	// Find process
	process, err := os.FindProcess(int(pid))
	if err != nil {
		return nil, fmt.Errorf("failed to find process: %w", err)
	}

	// Send signal
	switch signal {
	case "SIGKILL":
		err = process.Kill()
	case "SIGINT":
		err = process.Signal(os.Interrupt)
	default:
		err = process.Signal(os.Kill)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to send signal: %w", err)
	}

	return map[string]interface{}{
		"pid":    pid,
		"signal": signal,
		"status": "killed",
	}, nil
}
