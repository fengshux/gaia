// Package core provides the core engine
package core

import (
	"context"
	"fmt"
	"strings"

	"partner/internal/config"
	"partner/pkg/types"
	"partner/pkg/utils"
)

// PermissionManager handles permission checks
type PermissionManager struct {
	config config.PermissionsConfig
}

// NewPermissionManager creates a new permission manager
func NewPermissionManager(cfg config.PermissionsConfig) *PermissionManager {
	return &PermissionManager{
		config: cfg,
	}
}

// CheckPermission checks if an action is permitted
func (pm *PermissionManager) CheckPermission(action string, params map[string]interface{}) bool {
	switch pm.config.DefaultMode {
	case string(types.PermissionAuto):
		return true
	case string(types.PermissionDeny):
		return false
	default: // ask
		// For now, allow in auto mode, deny in deny mode
		// In interactive mode, this would prompt the user
		return pm.isAllowed(action, params)
	}
}

// isAllowed checks if an action is in the allowed list
func (pm *PermissionManager) isAllowed(action string, params map[string]interface{}) bool {
	// Check denied commands first
	for _, denied := range pm.config.DeniedCommands {
		if strings.Contains(action, denied) {
			return false
		}

		// Check if params contain denied command
		if cmd, ok := params["command"].(string); ok {
			if strings.Contains(cmd, denied) {
				return false
			}
		}
	}

	// Check allowed paths for file operations
	if action == "file_read" || action == "file_write" || action == "file_delete" {
		path, ok := params["path"].(string)
		if !ok {
			return false
		}

		// If no allowed paths defined, deny by default
		if len(pm.config.AllowedPaths) == 0 {
			return false
		}

		// Check if path is within allowed paths
		for _, allowed := range pm.config.AllowedPaths {
			if strings.HasPrefix(path, allowed) {
				return true
			}
		}

		return false
	}

	return true
}

// RequestPermission creates a permission request for user confirmation
func (pm *PermissionManager) RequestPermission(action string, params map[string]interface{}) *types.PermissionRequest {
	risk := pm.assessRisk(action, params)

	return &types.PermissionRequest{
		Action:      action,
		Resource:    pm.getResource(action, params),
		Params:      params,
		Risk:        risk,
		Description: pm.describe(action, params),
	}
}

// assessRisk assesses the risk level of an action
func (pm *PermissionManager) assessRisk(action string, params map[string]interface{}) string {
	// High risk actions
	highRisk := []string{"file_delete", "shell_execute", "email_send"}
	for _, hr := range highRisk {
		if action == hr {
			return "high"
		}
	}

	// Check for dangerous commands
	if cmd, ok := params["command"].(string); ok {
		dangerous := []string{"rm", "sudo", "chmod", "chown", "mkfs", "dd", ">", ">>"}
		for _, d := range dangerous {
			if strings.Contains(cmd, d) {
				return "high"
			}
		}
	}

	// Medium risk actions
	mediumRisk := []string{"file_write", "git_push", "git_reset"}
	for _, mr := range mediumRisk {
		if action == mr {
			return "medium"
		}
	}

	return "low"
}

// getResource extracts the resource being accessed
func (pm *PermissionManager) getResource(action string, params map[string]interface{}) string {
	switch action {
	case "file_read", "file_write", "file_delete":
		if path, ok := params["path"].(string); ok {
			return path
		}
	case "shell_execute":
		if cmd, ok := params["command"].(string); ok {
			return cmd
		}
	case "email_send":
		if to, ok := params["to"].(string); ok {
			return to
		}
	case "git_push", "git_pull", "git_commit":
		if repo, ok := params["repo"].(string); ok {
			return repo
		}
	}
	return "unknown"
}

// describe creates a human-readable description of the action
func (pm *PermissionManager) describe(action string, params map[string]interface{}) string {
	switch action {
	case "file_read":
		path, _ := params["path"].(string)
		return fmt.Sprintf("Read file: %s", path)
	case "file_write":
		path, _ := params["path"].(string)
		return fmt.Sprintf("Write to file: %s", path)
	case "file_delete":
		path, _ := params["path"].(string)
		return fmt.Sprintf("Delete file: %s", path)
	case "shell_execute":
		cmd, _ := params["command"].(string)
		return fmt.Sprintf("Execute command: %s", cmd)
	case "email_send":
		to, _ := params["to"].(string)
		return fmt.Sprintf("Send email to: %s", to)
	case "git_push":
		return "Push changes to remote repository"
	case "git_commit":
		msg, _ := params["message"].(string)
		return fmt.Sprintf("Commit changes: %s", msg)
	default:
		return fmt.Sprintf("Execute action: %s", action)
	}
}

// PermissionHandler handles permission requests interactively
type PermissionHandler func(ctx context.Context, req *types.PermissionRequest) (*types.PermissionResponse, error)

// DefaultPermissionHandler provides a default handler that always grants permission
func DefaultPermissionHandler(ctx context.Context, req *types.PermissionRequest) (*types.PermissionResponse, error) {
	return &types.PermissionResponse{
		Granted: true,
		Reason:  "Auto-approved",
	}, nil
}

// InteractivePermissionHandler creates a handler that prompts for user input
func InteractivePermissionHandler(prompt func(string) bool) PermissionHandler {
	return func(ctx context.Context, req *types.PermissionRequest) (*types.PermissionResponse, error) {
		msg := fmt.Sprintf("[Permission Request] %s (Risk: %s)\nAllow?", req.Description, req.Risk)
		granted := prompt(msg)
		return &types.PermissionResponse{
			Granted: granted,
			Reason:  "User decision",
		}, nil
	}
}

// Contains is a helper function (using utils package)
var Contains = utils.Contains[string]
