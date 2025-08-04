package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v73/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ListStarredRepositories creates a tool to list repositories that a user has starred.
func ListStarredRepositories(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_starred_repositories",
			mcp.WithDescription(t("TOOL_LIST_STARRED_REPOSITORIES_DESCRIPTION", "List repositories that a user has starred on GitHub. Returns at least 30 results per page by default, but can return more if specified using the perPage parameter (up to 100).")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_LIST_STARRED_REPOSITORIES_USER_TITLE", "List starred repositories"),
				ReadOnlyHint: ToBoolPtr(true),
			}),
			mcp.WithString("username",
				mcp.Required(),
				mcp.Description("GitHub username of the user whose starred repositories to list"),
			),
			mcp.WithString("sort",
				mcp.Description("Sort order for repositories. Can be 'created' (when the repository was starred) or 'updated' (when the repository was last pushed to). Default is 'created'."),
				mcp.Enum("created", "updated"),
			),
			mcp.WithString("direction",
				mcp.Description("Direction to sort repositories. Can be 'asc' or 'desc'. Default is 'desc'."),
				mcp.Enum("asc", "desc"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			username, err := RequiredParam[string](request, "username")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			sort, err := OptionalParam[string](request, "sort")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			direction, err := OptionalParam[string](request, "direction")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			// Validate sort parameter
			if sort != "" && sort != "created" && sort != "updated" {
				return mcp.NewToolResultError("invalid value for sort parameter"), nil
			}

			// Validate direction parameter
			if direction != "" && direction != "asc" && direction != "desc" {
				return mcp.NewToolResultError("invalid value for direction parameter"), nil
			}

			// Validate pagination parameters directly from request before using OptionalPaginationParams
			if pageVal, ok := request.GetArguments()["page"].(float64); ok {
				if pageVal <= 0 {
					return mcp.NewToolResultError("page must be greater than 0"), nil
				}
			}

			if perPageVal, ok := request.GetArguments()["perPage"].(float64); ok {
				if perPageVal <= 0 || perPageVal > 100 {
					return mcp.NewToolResultError("perPage must be between 1 and 100"), nil
				}
			}

			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Set default perPage to 30 if not provided (this should not happen due to validation above)
			perPage := pagination.PerPage
			if perPage == 0 {
				perPage = 30
			}

			// Set default page to 1 if not provided (this should not happen due to validation above)
			page := pagination.Page
			if page == 0 {
				page = 1
			}

			opts := &github.ActivityListStarredOptions{
				Sort:      sort,
				Direction: direction,
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: perPage,
				},
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			starredRepos, resp, err := client.Activity.ListStarred(ctx, username, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list starred repositories for user: %s: %w", username, err)
			}
			defer func() { _ = resp.Body.Close() }()

			r, err := json.Marshal(starredRepos)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}
