// Package tool provides built-in tools
package tool

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CodeRunTool runs code in various languages
type CodeRunTool struct{}

// NewCodeRunTool creates a new code run tool
func NewCodeRunTool() *CodeRunTool {
	return &CodeRunTool{}
}

func (t *CodeRunTool) Name() string {
	return "code_run"
}

func (t *CodeRunTool) Description() string {
	return "Run code in various programming languages"
}

func (t *CodeRunTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"language": StringParam{
				Type:        "string",
				Description: "Programming language",
				Enum:        []string{"python", "javascript", "go", "ruby", "bash"},
			},
			"code": StringParam{
				Type:        "string",
				Description: "The code to run",
			},
		},
		Required: []string{"language", "code"},
	}
}

func (t *CodeRunTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	language, ok := params["language"].(string)
	if !ok {
		return nil, fmt.Errorf("language parameter is required")
	}

	code, ok := params["code"].(string)
	if !ok {
		return nil, fmt.Errorf("code parameter is required")
	}

	var cmd *exec.Cmd

	switch strings.ToLower(language) {
	case "python":
		cmd = exec.CommandContext(ctx, "python3", "-c", code)
	case "javascript", "js":
		cmd = exec.CommandContext(ctx, "node", "-e", code)
	case "ruby":
		cmd = exec.CommandContext(ctx, "ruby", "-e", code)
	case "bash":
		cmd = exec.CommandContext(ctx, "bash", "-c", code)
	case "go":
		// For Go, we need to create a temp file
		// Simplified version - just run with go run
		cmd = exec.CommandContext(ctx, "go", "run", "-")
		cmd.Stdin = strings.NewReader(code)
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"error":    err.Error(),
			"output":   string(output),
			"language": language,
		}, nil
	}

	return map[string]interface{}{
		"output":   string(output),
		"language": language,
	}, nil
}

// CodeFormatTool formats code
type CodeFormatTool struct{}

// NewCodeFormatTool creates a new code format tool
func NewCodeFormatTool() *CodeFormatTool {
	return &CodeFormatTool{}
}

func (t *CodeFormatTool) Name() string {
	return "code_format"
}

func (t *CodeFormatTool) Description() string {
	return "Format code using standard formatters"
}

func (t *CodeFormatTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"language": StringParam{
				Type:        "string",
				Description: "Programming language",
				Enum:        []string{"go", "python", "javascript", "json"},
			},
			"code": StringParam{
				Type:        "string",
				Description: "The code to format",
			},
		},
		Required: []string{"language", "code"},
	}
}

func (t *CodeFormatTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	language, ok := params["language"].(string)
	if !ok {
		return nil, fmt.Errorf("language parameter is required")
	}

	code, ok := params["code"].(string)
	if !ok {
		return nil, fmt.Errorf("code parameter is required")
	}

	switch strings.ToLower(language) {
	case "go":
		// Use gofmt
		cmd := exec.CommandContext(ctx, "gofmt", "-s")
		cmd.Stdin = strings.NewReader(code)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("gofmt error: %w", err)
		}
		return string(output), nil

	case "python":
		// Use black or autopep8 if available
		cmd := exec.CommandContext(ctx, "python3", "-m", "autopep8", "-")
		cmd.Stdin = strings.NewReader(code)
		output, err := cmd.Output()
		if err != nil {
			// If autopep8 not available, return original
			return code, nil
		}
		return string(output), nil

	case "javascript", "js":
		// Use prettier if available
		cmd := exec.CommandContext(ctx, "prettier", "--stdin-filepath", "file.js")
		cmd.Stdin = strings.NewReader(code)
		output, err := cmd.Output()
		if err != nil {
			// If prettier not available, return original
			return code, nil
		}
		return string(output), nil

	case "json":
		// Use jq or python for JSON formatting
		cmd := exec.CommandContext(ctx, "python3", "-m", "json.tool")
		cmd.Stdin = strings.NewReader(code)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("json format error: %w", err)
		}
		return string(output), nil

	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

// CodeLintTool lints code
type CodeLintTool struct{}

// NewCodeLintTool creates a new code lint tool
func NewCodeLintTool() *CodeLintTool {
	return &CodeLintTool{}
}

func (t *CodeLintTool) Name() string {
	return "code_lint"
}

func (t *CodeLintTool) Description() string {
	return "Lint code using standard linters"
}

func (t *CodeLintTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"language": StringParam{
				Type:        "string",
				Description: "Programming language",
				Enum:        []string{"go", "python", "javascript"},
			},
			"code": StringParam{
				Type:        "string",
				Description: "The code to lint",
			},
		},
		Required: []string{"language", "code"},
	}
}

func (t *CodeLintTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	language, ok := params["language"].(string)
	if !ok {
		return nil, fmt.Errorf("language parameter is required")
	}

	code, ok := params["code"].(string)
	if !ok {
		return nil, fmt.Errorf("code parameter is required")
	}

	switch strings.ToLower(language) {
	case "go":
		// Use go vet
		cmd := exec.CommandContext(ctx, "go", "vet")
		cmd.Stdin = strings.NewReader(code)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return map[string]interface{}{
				"issues": string(output),
				"error":  err.Error(),
			}, nil
		}
		return map[string]interface{}{
			"issues": "",
			"status": "ok",
		}, nil

	case "python":
		// Use pylint or flake8 if available
		cmd := exec.CommandContext(ctx, "python3", "-m", "flake8", "-")
		cmd.Stdin = strings.NewReader(code)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return map[string]interface{}{
				"issues": string(output),
			}, nil
		}
		return map[string]interface{}{
			"issues": "",
			"status": "ok",
		}, nil

	case "javascript", "js":
		// Use eslint if available
		cmd := exec.CommandContext(ctx, "eslint", "--stdin")
		cmd.Stdin = strings.NewReader(code)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return map[string]interface{}{
				"issues": string(output),
			}, nil
		}
		return map[string]interface{}{
			"issues": "",
			"status": "ok",
		}, nil

	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

// CodeTestTool runs tests
type CodeTestTool struct{}

// NewCodeTestTool creates a new code test tool
func NewCodeTestTool() *CodeTestTool {
	return &CodeTestTool{}
}

func (t *CodeTestTool) Name() string {
	return "code_test"
}

func (t *CodeTestTool) Description() string {
	return "Run tests for a project"
}

func (t *CodeTestTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"language": StringParam{
				Type:        "string",
				Description: "Programming language",
				Enum:        []string{"go", "python", "javascript"},
			},
			"path": StringParam{
				Type:        "string",
				Description: "Path to the project or test file",
			},
			"pattern": StringParam{
				Type:        "string",
				Description: "Test pattern to match (optional)",
			},
		},
		Required: []string{"language", "path"},
	}
}

func (t *CodeTestTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	language, ok := params["language"].(string)
	if !ok {
		return nil, fmt.Errorf("language parameter is required")
	}

	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required")
	}

	pattern := ""
	if p, ok := params["pattern"].(string); ok {
		pattern = p
	}

	var cmd *exec.Cmd

	switch strings.ToLower(language) {
	case "go":
		args := []string{"test", "-v"}
		if pattern != "" {
			args = append(args, "-run", pattern)
		}
		args = append(args, path)
		cmd = exec.CommandContext(ctx, "go", args...)

	case "python":
		args := []string{"-m", "pytest", "-v"}
		if pattern != "" {
			args = append(args, "-k", pattern)
		}
		args = append(args, path)
		cmd = exec.CommandContext(ctx, "python3", args...)

	case "javascript", "js":
		cmd = exec.CommandContext(ctx, "npm", "test", "--", path)
		if pattern != "" {
			cmd = exec.CommandContext(ctx, "npm", "test", "--", "--grep", pattern, path)
		}

	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	output, err := cmd.CombinedOutput()

	return map[string]interface{}{
		"output": string(output),
		"error":  fmt.Sprintf("%v", err),
	}, nil
}
