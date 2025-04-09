package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SecurityAndAnalysis represents the security and analysis settings for a repository
type SecurityAndAnalysis struct {
	AdvancedSecurity struct {
		Status string `json:"status"`
	} `json:"advanced_security"`
	SecretScanning struct {
		Status string `json:"status"`
	} `json:"secret_scanning"`
	SecretScanningPushProtection struct {
		Status string `json:"status"`
	} `json:"secret_scanning_push_protection"`
}

// GetSecuritySettings retrieves security settings for a repository
func GetSecuritySettings(client *github.Client, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_security_settings",
		mcp.WithDescription(t("TOOL_GET_SECURITY_SETTINGS_DESCRIPTION", "Get security settings for a repository")),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description(t("PARAM_OWNER_DESCRIPTION", "Repository owner")),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description(t("PARAM_REPO_DESCRIPTION", "Repository name")),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner, ok := request.Params.Arguments["owner"].(string)
		if !ok {
			return nil, fmt.Errorf("missing required parameter: owner")
		}

		repo, ok := request.Params.Arguments["repo"].(string)
		if !ok {
			return nil, fmt.Errorf("missing required parameter: repo")
		}

		repository, _, err := client.Repositories.Get(ctx, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get repository settings: %w", err)
		}

		response, err := json.Marshal(repository.SecurityAndAnalysis)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		return mcp.NewToolResultText(string(response)), nil
	}
}

// UpdateSecuritySettings updates security settings for a repository
func UpdateSecuritySettings(client *github.Client, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("update_security_settings",
		mcp.WithDescription(t("TOOL_UPDATE_SECURITY_SETTINGS_DESCRIPTION", "Update security settings for a repository")),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description(t("PARAM_OWNER_DESCRIPTION", "Repository owner")),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description(t("PARAM_REPO_DESCRIPTION", "Repository name")),
		),
		mcp.WithObject("settings",
			mcp.Required(),
			mcp.Description(t("PARAM_SETTINGS_DESCRIPTION", "Security settings to update")),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner, ok := request.Params.Arguments["owner"].(string)
		if !ok {
			return nil, fmt.Errorf("missing required parameter: owner")
		}

		repo, ok := request.Params.Arguments["repo"].(string)
		if !ok {
			return nil, fmt.Errorf("missing required parameter: repo")
		}

		settings, ok := request.Params.Arguments["settings"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing required parameter: settings")
		}

		// Get current repository settings
		repository, _, err := client.Repositories.Get(ctx, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get repository settings: %w", err)
		}

		// Initialize security settings if nil
		if repository.SecurityAndAnalysis == nil {
			repository.SecurityAndAnalysis = &github.SecurityAndAnalysis{}
		}

		// Update vulnerability alerts if specified
		if vulnerabilityAlerts, ok := settings["vulnerability_alerts"].(bool); ok {
			if repository.SecurityAndAnalysis.AdvancedSecurity == nil {
				repository.SecurityAndAnalysis.AdvancedSecurity = &github.AdvancedSecurity{}
			}
			if vulnerabilityAlerts {
				repository.SecurityAndAnalysis.AdvancedSecurity.Status = github.Ptr("enabled")
			} else {
				repository.SecurityAndAnalysis.AdvancedSecurity.Status = github.Ptr("disabled")
			}
		}

		// Update other security settings
		settingsJSON, err := json.Marshal(settings)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal settings: %w", err)
		}

		var securitySettings github.SecurityAndAnalysis
		if err := json.Unmarshal(settingsJSON, &securitySettings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}

		// Merge the new settings with existing ones
		if securitySettings.AdvancedSecurity != nil {
			if repository.SecurityAndAnalysis.AdvancedSecurity == nil || repository.SecurityAndAnalysis.AdvancedSecurity.Status == "" {
				repository.SecurityAndAnalysis.AdvancedSecurity = securitySettings.AdvancedSecurity
			}
		}
		if securitySettings.SecretScanning != nil {
			repository.SecurityAndAnalysis.SecretScanning = securitySettings.SecretScanning
		}
		if securitySettings.SecretScanningPushProtection != nil {
			repository.SecurityAndAnalysis.SecretScanningPushProtection = securitySettings.SecretScanningPushProtection
		}

		// Update the repository
		updatedRepo, _, err := client.Repositories.Edit(ctx, owner, repo, &github.Repository{
			SecurityAndAnalysis: repository.SecurityAndAnalysis,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update repository settings: %w", err)
		}

		// Return complete security settings
		response, err := json.Marshal(updatedRepo.SecurityAndAnalysis)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		return mcp.NewToolResultText(string(response)), nil
	}
}

// GetDependabotSecurityUpdatesStatus checks if Dependabot security updates are enabled
func GetDependabotSecurityUpdatesStatus(client *github.Client, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_dependabot_security_updates_status",
		mcp.WithDescription(t("TOOL_GET_DEPENDABOT_SECURITY_UPDATES_STATUS_DESCRIPTION", "Check if Dependabot security updates are enabled for a repository")),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description(t("PARAM_OWNER_DESCRIPTION", "Repository owner")),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description(t("PARAM_REPO_DESCRIPTION", "Repository name")),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner, ok := request.Params.Arguments["owner"].(string)
		if !ok {
			return nil, fmt.Errorf("missing required parameter: owner")
		}

		repo, ok := request.Params.Arguments["repo"].(string)
		if !ok {
			return nil, fmt.Errorf("missing required parameter: repo")
		}

		status, _, err := client.Repositories.GetAutomatedSecurityFixes(ctx, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get Dependabot security updates status: %w", err)
		}

		response, err := json.Marshal(status)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		return mcp.NewToolResultText(string(response)), nil
	}
}

// EnableDependabotSecurityUpdates and DisableDependabotSecurityUpdates are currently disabled.
// Issue: There is a discrepancy in GitHub's API behavior regarding Dependabot security updates:
// 1. Public repositories should have Dependabot alerts enabled by default
// 2. However, the API still requires explicit enabling of vulnerability alerts
// 3. This creates a confusing user experience where the system says one thing but behaves differently
// 4. The functionality needs to be investigated and fixed before being re-enabled
// See: https://github.com/github/github-mcp-server/issues/176

// EnableDependabotSecurityUpdates enables Dependabot security updates for a repository
// func EnableDependabotSecurityUpdates(client *github.Client, t translations.TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
// 	return mcp.NewTool("enable_dependabot_security_updates",
// 		mcp.WithDescription(t("TOOL_ENABLE_DEPENDABOT_SECURITY_UPDATES_DESCRIPTION", "Enable Dependabot security updates for a repository")),
// 		mcp.WithString("owner",
// 			mcp.Required(),
// 			mcp.Description(t("PARAM_OWNER_DESCRIPTION", "Repository owner")),
// 		),
// 		mcp.WithString("repo",
// 			mcp.Required(),
// 			mcp.Description(t("PARAM_REPO_DESCRIPTION", "Repository name")),
// 		),
// 	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
// 		return nil, fmt.Errorf("this functionality is currently disabled due to GitHub API behavior discrepancy")
// 	}
// }

// DisableDependabotSecurityUpdates disables Dependabot security updates for a repository
// func DisableDependabotSecurityUpdates(client *github.Client, t translations.TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
// 	return mcp.NewTool("disable_dependabot_security_updates",
// 		mcp.WithDescription(t("TOOL_DISABLE_DEPENDABOT_SECURITY_UPDATES_DESCRIPTION", "Disable Dependabot security updates for a repository")),
// 		mcp.WithString("owner",
// 			mcp.Required(),
// 			mcp.Description(t("PARAM_OWNER_DESCRIPTION", "Repository owner")),
// 		),
// 		mcp.WithString("repo",
// 			mcp.Required(),
// 			mcp.Description(t("PARAM_REPO_DESCRIPTION", "Repository name")),
// 		),
// 	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
// 		return nil, fmt.Errorf("this functionality is currently disabled due to GitHub API behavior discrepancy")
// 	}
// }