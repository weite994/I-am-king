package github

import (
	"context"
	"fmt"

	"github.com/github/github-mcp-server/pkg/toolsets"
	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shurcooL/githubv4"
)

type GetClientFn func(context.Context) (*github.Client, error)
type GetGQLClientFn func(context.Context) (*githubv4.Client, error)

var DefaultTools = []string{"all"}

// extractAuthTokenFromRequest is a helper function that extracts the auth_token parameter
// from an MCP request and returns a context with the token injected
func extractAuthTokenFromRequest(ctx context.Context, request mcp.CallToolRequest) (context.Context, error) {
	token, err := requiredParam[string](request, "auth_token")
	if err != nil {
		return nil, err
	}
	
	return context.WithValue(ctx, "auth_token", token), nil
}

// wrapToolHandlerWithAuth wraps a tool handler to extract auth_token from the request
// and inject it into the context before calling the original handler
func wrapToolHandlerWithAuth(handler server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract auth token and inject into context
		ctxWithAuth, err := extractAuthTokenFromRequest(ctx, request)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("authentication error: %s", err.Error())), nil
		}
		
		// Call the original handler with the context containing the auth token
		return handler(ctxWithAuth, request)
	}
}

// createMultiUserTool creates a tool with auth token parameter and wraps the handler
func createMultiUserTool(tool mcp.Tool, handler server.ToolHandlerFunc) server.ServerTool {
	// Add auth_token parameter to the tool schema
	if tool.InputSchema.Properties == nil {
		tool.InputSchema.Properties = make(map[string]interface{})
	}
	tool.InputSchema.Properties["auth_token"] = map[string]interface{}{
		"type":        "string",
		"description": "GitHub Personal Access Token for authentication",
	}
	tool.InputSchema.Required = append(tool.InputSchema.Required, "auth_token")
	
	// Wrap the handler to extract auth token
	wrappedHandler := wrapToolHandlerWithAuth(handler)
	
	return toolsets.NewServerTool(tool, wrappedHandler)
}

// wrapToolFunc is a helper that takes a tool function and wraps it for multi-user support
func wrapToolFunc(toolFunc func() (mcp.Tool, server.ToolHandlerFunc)) server.ServerTool {
	tool, handler := toolFunc()
	return createMultiUserTool(tool, handler)
}

func InitToolsets(passedToolsets []string, readOnly bool, getClient GetClientFn, getGQLClient GetGQLClientFn, t translations.TranslationHelperFunc) (*toolsets.ToolsetGroup, error) {
	// Create a new toolset group
	tsg := toolsets.NewToolsetGroup(readOnly)

	// Define all available features with their default state (disabled)
	// Create toolsets
	repos := toolsets.NewToolset("repos", "GitHub Repository related tools").
		AddReadTools(
			toolsets.NewServerTool(SearchRepositories(getClient, t)),
			toolsets.NewServerTool(GetFileContents(getClient, t)),
			toolsets.NewServerTool(ListCommits(getClient, t)),
			toolsets.NewServerTool(SearchCode(getClient, t)),
			toolsets.NewServerTool(GetCommit(getClient, t)),
			toolsets.NewServerTool(ListBranches(getClient, t)),
			toolsets.NewServerTool(ListTags(getClient, t)),
			toolsets.NewServerTool(GetTag(getClient, t)),
		).
		AddWriteTools(
			toolsets.NewServerTool(CreateOrUpdateFile(getClient, t)),
			toolsets.NewServerTool(CreateRepository(getClient, t)),
			toolsets.NewServerTool(ForkRepository(getClient, t)),
			toolsets.NewServerTool(CreateBranch(getClient, t)),
			toolsets.NewServerTool(PushFiles(getClient, t)),
			toolsets.NewServerTool(DeleteFile(getClient, t)),
		)
	issues := toolsets.NewToolset("issues", "GitHub Issues related tools").
		AddReadTools(
			toolsets.NewServerTool(GetIssue(getClient, t)),
			toolsets.NewServerTool(SearchIssues(getClient, t)),
			toolsets.NewServerTool(ListIssues(getClient, t)),
			toolsets.NewServerTool(GetIssueComments(getClient, t)),
		).
		AddWriteTools(
			toolsets.NewServerTool(CreateIssue(getClient, t)),
			toolsets.NewServerTool(AddIssueComment(getClient, t)),
			toolsets.NewServerTool(UpdateIssue(getClient, t)),
			toolsets.NewServerTool(AssignCopilotToIssue(getGQLClient, t)),
		)
	users := toolsets.NewToolset("users", "GitHub User related tools").
		AddReadTools(
			toolsets.NewServerTool(SearchUsers(getClient, t)),
		)
	pullRequests := toolsets.NewToolset("pull_requests", "GitHub Pull Request related tools").
		AddReadTools(
			toolsets.NewServerTool(GetPullRequest(getClient, t)),
			toolsets.NewServerTool(ListPullRequests(getClient, t)),
			toolsets.NewServerTool(GetPullRequestFiles(getClient, t)),
			toolsets.NewServerTool(GetPullRequestStatus(getClient, t)),
			toolsets.NewServerTool(GetPullRequestComments(getClient, t)),
			toolsets.NewServerTool(GetPullRequestReviews(getClient, t)),
			toolsets.NewServerTool(GetPullRequestDiff(getClient, t)),
		).
		AddWriteTools(
			toolsets.NewServerTool(MergePullRequest(getClient, t)),
			toolsets.NewServerTool(UpdatePullRequestBranch(getClient, t)),
			toolsets.NewServerTool(CreatePullRequest(getClient, t)),
			toolsets.NewServerTool(UpdatePullRequest(getClient, t)),
			toolsets.NewServerTool(RequestCopilotReview(getClient, t)),

			// Reviews
			toolsets.NewServerTool(CreateAndSubmitPullRequestReview(getGQLClient, t)),
			toolsets.NewServerTool(CreatePendingPullRequestReview(getGQLClient, t)),
			toolsets.NewServerTool(AddPullRequestReviewCommentToPendingReview(getGQLClient, t)),
			toolsets.NewServerTool(SubmitPendingPullRequestReview(getGQLClient, t)),
			toolsets.NewServerTool(DeletePendingPullRequestReview(getGQLClient, t)),
		)
	codeSecurity := toolsets.NewToolset("code_security", "Code security related tools, such as GitHub Code Scanning").
		AddReadTools(
			toolsets.NewServerTool(GetCodeScanningAlert(getClient, t)),
			toolsets.NewServerTool(ListCodeScanningAlerts(getClient, t)),
		)
	secretProtection := toolsets.NewToolset("secret_protection", "Secret protection related tools, such as GitHub Secret Scanning").
		AddReadTools(
			toolsets.NewServerTool(GetSecretScanningAlert(getClient, t)),
			toolsets.NewServerTool(ListSecretScanningAlerts(getClient, t)),
		)

	notifications := toolsets.NewToolset("notifications", "GitHub Notifications related tools").
		AddReadTools(
			toolsets.NewServerTool(ListNotifications(getClient, t)),
			toolsets.NewServerTool(GetNotificationDetails(getClient, t)),
		).
		AddWriteTools(
			toolsets.NewServerTool(DismissNotification(getClient, t)),
			toolsets.NewServerTool(MarkAllNotificationsRead(getClient, t)),
			toolsets.NewServerTool(ManageNotificationSubscription(getClient, t)),
			toolsets.NewServerTool(ManageRepositoryNotificationSubscription(getClient, t)),
		)

	// Keep experiments alive so the system doesn't error out when it's always enabled
	experiments := toolsets.NewToolset("experiments", "Experimental features that are not considered stable yet")

	// Add toolsets to the group
	tsg.AddToolset(repos)
	tsg.AddToolset(issues)
	tsg.AddToolset(users)
	tsg.AddToolset(pullRequests)
	tsg.AddToolset(codeSecurity)
	tsg.AddToolset(secretProtection)
	tsg.AddToolset(notifications)
	tsg.AddToolset(experiments)
	// Enable the requested features

	if err := tsg.EnableToolsets(passedToolsets); err != nil {
		return nil, err
	}

	return tsg, nil
}

// InitMultiUserToolsets creates toolsets that support multiple users by extracting
// auth tokens from each request and adding auth_token parameter to all tools
func InitMultiUserToolsets(passedToolsets []string, readOnly bool, getClient GetClientFn, getGQLClient GetGQLClientFn, t translations.TranslationHelperFunc) (*toolsets.ToolsetGroup, error) {
	// Create a new toolset group
	tsg := toolsets.NewToolsetGroup(readOnly)

	// Create all tool definitions with auth token support
	repos := toolsets.NewToolset("repos", "GitHub Repository related tools").
		AddReadTools(
			createMultiUserTool(SearchRepositories(getClient, t)),
			createMultiUserTool(GetFileContents(getClient, t)),
			createMultiUserTool(ListCommits(getClient, t)),
			createMultiUserTool(SearchCode(getClient, t)),
			createMultiUserTool(GetCommit(getClient, t)),
			createMultiUserTool(ListBranches(getClient, t)),
			createMultiUserTool(ListTags(getClient, t)),
			createMultiUserTool(GetTag(getClient, t)),
		).
		AddWriteTools(
			createMultiUserTool(CreateOrUpdateFile(getClient, t)),
			createMultiUserTool(CreateRepository(getClient, t)),
			createMultiUserTool(ForkRepository(getClient, t)),
			createMultiUserTool(CreateBranch(getClient, t)),
			createMultiUserTool(PushFiles(getClient, t)),
			createMultiUserTool(DeleteFile(getClient, t)),
		)
	issues := toolsets.NewToolset("issues", "GitHub Issues related tools").
		AddReadTools(
			createMultiUserTool(GetIssue(getClient, t)),
			createMultiUserTool(SearchIssues(getClient, t)),
			createMultiUserTool(ListIssues(getClient, t)),
			createMultiUserTool(GetIssueComments(getClient, t)),
		).
		AddWriteTools(
			createMultiUserTool(CreateIssue(getClient, t)),
			createMultiUserTool(AddIssueComment(getClient, t)),
			createMultiUserTool(UpdateIssue(getClient, t)),
			createMultiUserTool(AssignCopilotToIssue(getGQLClient, t)),
		)
	users := toolsets.NewToolset("users", "GitHub User related tools").
		AddReadTools(
			createMultiUserTool(SearchUsers(getClient, t)),
		)
	pullRequests := toolsets.NewToolset("pull_requests", "GitHub Pull Request related tools").
		AddReadTools(
			createMultiUserTool(GetPullRequest(getClient, t)),
			createMultiUserTool(ListPullRequests(getClient, t)),
			createMultiUserTool(GetPullRequestFiles(getClient, t)),
			createMultiUserTool(GetPullRequestStatus(getClient, t)),
			createMultiUserTool(GetPullRequestComments(getClient, t)),
			createMultiUserTool(GetPullRequestReviews(getClient, t)),
			createMultiUserTool(GetPullRequestDiff(getClient, t)),
		).
		AddWriteTools(
			createMultiUserTool(MergePullRequest(getClient, t)),
			createMultiUserTool(UpdatePullRequestBranch(getClient, t)),
			createMultiUserTool(CreatePullRequest(getClient, t)),
			createMultiUserTool(UpdatePullRequest(getClient, t)),
			createMultiUserTool(RequestCopilotReview(getClient, t)),

			// Reviews
			createMultiUserTool(CreateAndSubmitPullRequestReview(getGQLClient, t)),
			createMultiUserTool(CreatePendingPullRequestReview(getGQLClient, t)),
			createMultiUserTool(AddPullRequestReviewCommentToPendingReview(getGQLClient, t)),
			createMultiUserTool(SubmitPendingPullRequestReview(getGQLClient, t)),
			createMultiUserTool(DeletePendingPullRequestReview(getGQLClient, t)),
		)
	codeSecurity := toolsets.NewToolset("code_security", "Code security related tools, such as GitHub Code Scanning").
		AddReadTools(
			createMultiUserTool(GetCodeScanningAlert(getClient, t)),
			createMultiUserTool(ListCodeScanningAlerts(getClient, t)),
		)
	secretProtection := toolsets.NewToolset("secret_protection", "Secret protection related tools, such as GitHub Secret Scanning").
		AddReadTools(
			createMultiUserTool(GetSecretScanningAlert(getClient, t)),
			createMultiUserTool(ListSecretScanningAlerts(getClient, t)),
		)

	notifications := toolsets.NewToolset("notifications", "GitHub Notifications related tools").
		AddReadTools(
			createMultiUserTool(ListNotifications(getClient, t)),
			createMultiUserTool(GetNotificationDetails(getClient, t)),
		).
		AddWriteTools(
			createMultiUserTool(DismissNotification(getClient, t)),
			createMultiUserTool(MarkAllNotificationsRead(getClient, t)),
			createMultiUserTool(ManageNotificationSubscription(getClient, t)),
			createMultiUserTool(ManageRepositoryNotificationSubscription(getClient, t)),
		)

	// Keep experiments alive so the system doesn't error out when it's always enabled
	experiments := toolsets.NewToolset("experiments", "Experimental features that are not considered stable yet")

	// Add toolsets to the group
	tsg.AddToolset(repos)
	tsg.AddToolset(issues)
	tsg.AddToolset(users)
	tsg.AddToolset(pullRequests)
	tsg.AddToolset(codeSecurity)
	tsg.AddToolset(secretProtection)
	tsg.AddToolset(notifications)
	tsg.AddToolset(experiments)
	// Enable the requested features

	if err := tsg.EnableToolsets(passedToolsets); err != nil {
		return nil, err
	}

	return tsg, nil
}

func InitContextToolset(getClient GetClientFn, t translations.TranslationHelperFunc) *toolsets.Toolset {
	// Create a new context toolset
	contextTools := toolsets.NewToolset("context", "Tools that provide context about the current user and GitHub context you are operating in").
		AddReadTools(
			toolsets.NewServerTool(GetMe(getClient, t)),
		)
	contextTools.Enabled = true
	return contextTools
}

// InitMultiUserContextToolset creates a context toolset that supports multiple users
func InitMultiUserContextToolset(getClient GetClientFn, t translations.TranslationHelperFunc) *toolsets.Toolset {
	// Create a new context toolset
	contextTools := toolsets.NewToolset("context", "Tools that provide context about the current user and GitHub context you are operating in").
		AddReadTools(
			createMultiUserTool(GetMe(getClient, t)),
		)
	contextTools.Enabled = true
	return contextTools
}

// InitDynamicToolset creates a dynamic toolset that can be used to enable other toolsets, and so requires the server and toolset group as arguments
func InitDynamicToolset(s *server.MCPServer, tsg *toolsets.ToolsetGroup, t translations.TranslationHelperFunc) *toolsets.Toolset {
	// Create a new dynamic toolset
	// Need to add the dynamic toolset last so it can be used to enable other toolsets
	dynamicToolSelection := toolsets.NewToolset("dynamic", "Discover GitHub MCP tools that can help achieve tasks by enabling additional sets of tools, you can control the enablement of any toolset to access its tools when this toolset is enabled.").
		AddReadTools(
			toolsets.NewServerTool(ListAvailableToolsets(tsg, t)),
			toolsets.NewServerTool(GetToolsetsTools(tsg, t)),
			toolsets.NewServerTool(EnableToolset(s, tsg, t)),
		)

	dynamicToolSelection.Enabled = true
	return dynamicToolSelection
}

func toBoolPtr(b bool) *bool {
	return &b
}
