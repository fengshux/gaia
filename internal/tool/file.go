// Package tool provides built-in tools
package tool

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"assistant/pkg/utils"
)

// FileReadTool reads a file
type FileReadTool struct{}

// NewFileReadTool creates a new file read tool
func NewFileReadTool() *FileReadTool {
	return &FileReadTool{}
}

func (t *FileReadTool) Name() string {
	return "file_read"
}

func (t *FileReadTool) Description() string {
	return "Read the contents of a file"
}

func (t *FileReadTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"path": StringParam{
				Type:        "string",
				Description: "The path to the file to read",
			},
			"offset": NumberParam{
				Type:        "number",
				Description: "Line number to start reading from (optional)",
			},
			"limit": NumberParam{
				Type:        "number",
				Description: "Maximum number of lines to read (optional)",
			},
		},
		Required: []string{"path"},
	}
}

func (t *FileReadTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required")
	}

	path = utils.ExpandPath(path)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// FileWriteTool writes to a file
type FileWriteTool struct{}

// NewFileWriteTool creates a new file write tool
func NewFileWriteTool() *FileWriteTool {
	return &FileWriteTool{}
}

func (t *FileWriteTool) Name() string {
	return "file_write"
}

func (t *FileWriteTool) Description() string {
	return "Write content to a file"
}

func (t *FileWriteTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"path": StringParam{
				Type:        "string",
				Description: "The path to the file to write",
			},
			"content": StringParam{
				Type:        "string",
				Description: "The content to write to the file",
			},
			"append": BooleanParam{
				Type:        "boolean",
				Description: "Whether to append to the file (default: false)",
			},
		},
		Required: []string{"path", "content"},
	}
}

func (t *FileWriteTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required")
	}

	content, ok := params["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content parameter is required")
	}

	append := false
	if v, ok := params["append"].(bool); ok {
		append = v
	}

	path = utils.ExpandPath(path)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := utils.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	flag := os.O_WRONLY | os.O_CREATE
	if append {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	file, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

// FileDeleteTool deletes a file
type FileDeleteTool struct{}

// NewFileDeleteTool creates a new file delete tool
func NewFileDeleteTool() *FileDeleteTool {
	return &FileDeleteTool{}
}

func (t *FileDeleteTool) Name() string {
	return "file_delete"
}

func (t *FileDeleteTool) Description() string {
	return "Delete a file"
}

func (t *FileDeleteTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"path": StringParam{
				Type:        "string",
				Description: "The path to the file to delete",
			},
		},
		Required: []string{"path"},
	}
}

func (t *FileDeleteTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required")
	}

	path = utils.ExpandPath(path)

	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return fmt.Sprintf("Successfully deleted %s", path), nil
}

// FileListTool lists files in a directory
type FileListTool struct{}

// NewFileListTool creates a new file list tool
func NewFileListTool() *FileListTool {
	return &FileListTool{}
}

func (t *FileListTool) Name() string {
	return "file_list"
}

func (t *FileListTool) Description() string {
	return "List files in a directory"
}

func (t *FileListTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"path": StringParam{
				Type:        "string",
				Description: "The path to the directory to list (default: current directory)",
			},
			"pattern": StringParam{
				Type:        "string",
				Description: "Glob pattern to filter files (optional)",
			},
			"recursive": BooleanParam{
				Type:        "boolean",
				Description: "Whether to list recursively (default: false)",
			},
		},
	}
}

func (t *FileListTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	path := "."
	if p, ok := params["path"].(string); ok {
		path = p
	}

	path = utils.ExpandPath(path)

	pattern := "*"
	if p, ok := params["pattern"].(string); ok {
		pattern = p
	}

	recursive := false
	if r, ok := params["recursive"].(bool); ok {
		recursive = r
	}

	var files []string

	if recursive {
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				matched, _ := filepath.Match(pattern, filepath.Base(filePath))
				if matched {
					files = append(files, filePath)
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			matched, _ := filepath.Match(pattern, entry.Name())
			if matched {
				files = append(files, filepath.Join(path, entry.Name()))
			}
		}
	}

	return strings.Join(files, "\n"), nil
}

// FileExistsTool checks if a file exists
type FileExistsTool struct{}

// NewFileExistsTool creates a new file exists tool
func NewFileExistsTool() *FileExistsTool {
	return &FileExistsTool{}
}

func (t *FileExistsTool) Name() string {
	return "file_exists"
}

func (t *FileExistsTool) Description() string {
	return "Check if a file or directory exists"
}

func (t *FileExistsTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"path": StringParam{
				Type:        "string",
				Description: "The path to check",
			},
		},
		Required: []string{"path"},
	}
}

func (t *FileExistsTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required")
	}

	path = utils.ExpandPath(path)

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{
				"exists": false,
				"path":   path,
			}, nil
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return map[string]interface{}{
		"exists":  true,
		"path":    path,
		"isDir":   info.IsDir(),
		"size":    info.Size(),
		"modTime": info.ModTime(),
	}, nil
}
