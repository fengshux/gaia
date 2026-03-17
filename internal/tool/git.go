// Package tool provides built-in tools
package tool

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitStatusTool checks git status
type GitStatusTool struct{}

// NewGitStatusTool creates a new git status tool
func NewGitStatusTool() *GitStatusTool {
	return &GitStatusTool{}
}

func (t *GitStatusTool) Name() string {
	return "git_status"
}

func (t *GitStatusTool) Description() string {
	return "Check git repository status"
}

func (t *GitStatusTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"path": StringParam{
				Type:        "string",
				Description: "Path to git repository (default: current directory)",
			},
		},
	}
}

func (t *GitStatusTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	path := "."
	if p, ok := params["path"].(string); ok {
		path = p
	}

	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}

	// Parse status
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var modified, added, deleted, untracked []string

	for _, line := range lines {
		if line == "" {
			continue
		}

		status := strings.TrimSpace(line[:2])
		file := strings.TrimSpace(line[3:])

		switch {
		case strings.Contains(status, "M"):
			modified = append(modified, file)
		case strings.Contains(status, "A"), strings.Contains(status, "?"):
			added = append(added, file)
		case strings.Contains(status, "D"):
			deleted = append(deleted, file)
		default:
			untracked = append(untracked, file)
		}
	}

	return map[string]interface{}{
		"modified":  modified,
		"added":     added,
		"deleted":   deleted,
		"untracked": untracked,
		"clean":     len(lines) == 0 || (len(lines) == 1 && lines[0] == ""),
	}, nil
}

// GitLogTool shows git log
type GitLogTool struct{}

// NewGitLogTool creates a new git log tool
func NewGitLogTool() *GitLogTool {
	return &GitLogTool{}
}

func (t *GitLogTool) Name() string {
	return "git_log"
}

func (t *GitLogTool) Description() string {
	return "Show git commit history"
}

func (t *GitLogTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"path": StringParam{
				Type:        "string",
				Description: "Path to git repository",
			},
			"count": NumberParam{
				Type:        "number",
				Description: "Number of commits to show (default: 10)",
			},
			"oneline": BooleanParam{
				Type:        "boolean",
				Description: "Show one line per commit (default: true)",
			},
		},
	}
}

func (t *GitLogTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	path := "."
	if p, ok := params["path"].(string); ok {
		path = p
	}

	count := 10
	if c, ok := params["count"].(float64); ok {
		count = int(c)
	}

	args := []string{"log", fmt.Sprintf("-%d", count)}
	if oneline, ok := params["oneline"].(bool); !ok || oneline {
		args = append(args, "--oneline")
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	return string(output), nil
}

// GitDiffTool shows git diff
type GitDiffTool struct{}

// NewGitDiffTool creates a new git diff tool
func NewGitDiffTool() *GitDiffTool {
	return &GitDiffTool{}
}

func (t *GitDiffTool) Name() string {
	return "git_diff"
}

func (t *GitDiffTool) Description() string {
	return "Show git diff"
}

func (t *GitDiffTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"path": StringParam{
				Type:        "string",
				Description: "Path to git repository",
			},
			"file": StringParam{
				Type:        "string",
				Description: "Specific file to diff (optional)",
			},
			"staged": BooleanParam{
				Type:        "boolean",
				Description: "Show staged changes (default: false)",
			},
		},
	}
}

func (t *GitDiffTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	path := "."
	if p, ok := params["path"].(string); ok {
		path = p
	}

	args := []string{"diff"}
	if staged, ok := params["staged"].(bool); ok && staged {
		args = append(args, "--staged")
	}
	if file, ok := params["file"].(string); ok {
		args = append(args, file)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	return string(output), nil
}

// GitCommitTool commits changes
type GitCommitTool struct{}

// NewGitCommitTool creates a new git commit tool
func NewGitCommitTool() *GitCommitTool {
	return &GitCommitTool{}
}

func (t *GitCommitTool) Name() string {
	return "git_commit"
}

func (t *GitCommitTool) Description() string {
	return "Commit changes to git"
}

func (t *GitCommitTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"path": StringParam{
				Type:        "string",
				Description: "Path to git repository",
			},
			"message": StringParam{
				Type:        "string",
				Description: "Commit message",
			},
			"files": ArrayParam{
				Type:        "array",
				Description: "Files to commit (optional, default: all staged)",
				Items:       StringParam{Type: "string"},
			},
		},
		Required: []string{"message"},
	}
}

func (t *GitCommitTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	path := "."
	if p, ok := params["path"].(string); ok {
		path = p
	}

	message, ok := params["message"].(string)
	if !ok {
		return nil, fmt.Errorf("message parameter is required")
	}

	// Add files if specified
	if files, ok := params["files"].([]interface{}); ok {
		fileArgs := []string{"add"}
		for _, f := range files {
			fileArgs = append(fileArgs, fmt.Sprintf("%v", f))
		}
		cmd := exec.CommandContext(ctx, "git", fileArgs...)
		cmd.Dir = path
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("git add failed: %w", err)
		}
	}

	// Commit
	cmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git commit failed: %w", err)
	}

	return string(output), nil
}

// GitBranchTool manages branches
type GitBranchTool struct{}

// NewGitBranchTool creates a new git branch tool
func NewGitBranchTool() *GitBranchTool {
	return &GitBranchTool{}
}

func (t *GitBranchTool) Name() string {
	return "git_branch"
}

func (t *GitBranchTool) Description() string {
	return "List, create, or delete git branches"
}

func (t *GitBranchTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"path": StringParam{
				Type:        "string",
				Description: "Path to git repository",
			},
			"action": StringParam{
				Type:        "string",
				Description: "Action to perform",
				Enum:        []string{"list", "create", "delete", "current"},
			},
			"name": StringParam{
				Type:        "string",
				Description: "Branch name (for create/delete)",
			},
		},
	}
}

func (t *GitBranchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	path := "."
	if p, ok := params["path"].(string); ok {
		path = p
	}

	action := "list"
	if a, ok := params["action"].(string); ok {
		action = a
	}

	switch action {
	case "list":
		cmd := exec.CommandContext(ctx, "git", "branch", "-a")
		cmd.Dir = path
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("git branch failed: %w", err)
		}
		return string(output), nil

	case "current":
		cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
		cmd.Dir = path
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("git branch failed: %w", err)
		}
		return strings.TrimSpace(string(output)), nil

	case "create":
		name, ok := params["name"].(string)
		if !ok {
			return nil, fmt.Errorf("name parameter is required for create")
		}
		cmd := exec.CommandContext(ctx, "git", "checkout", "-b", name)
		cmd.Dir = path
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("git branch create failed: %w", err)
		}
		return string(output), nil

	case "delete":
		name, ok := params["name"].(string)
		if !ok {
			return nil, fmt.Errorf("name parameter is required for delete")
		}
		cmd := exec.CommandContext(ctx, "git", "branch", "-d", name)
		cmd.Dir = path
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("git branch delete failed: %w", err)
		}
		return string(output), nil

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}
