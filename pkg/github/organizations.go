package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ListCommits creates a tool to get commits of a branch in a repository.
func ListRepositories(client *github.Client, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_repositories",
			mcp.WithDescription(t("TOOL_LIST_REPOSITORIES_DESCRIPTION", "Get list of repositories in a GitHub organization")),
			mcp.WithString("org",
				mcp.Required(),
				mcp.Description("Organization name"),
			),
			mcp.WithString("type",
				mcp.Description("Type of repositories to list. Possible values are: all, public, private, forks, sources, member. Default is 'all'."),
			),
			mcp.WithString("sort",
				mcp.Description("How to sort the repository list. Can be one of created, updated, pushed, full_name. Default is 'created'"),
			),
			mcp.WithString("direction",
				mcp.Description("Direction in which to sort repositories. Can be one of asc or desc. Default when using full_name: asc; otherwise desc."),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			org, err := requiredParam[string](request, "org")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.RepositoryListByOrgOptions{
				ListOptions: github.ListOptions{
					Page:    pagination.page,
					PerPage: pagination.perPage,
				},
			}

			repo_type, err := OptionalParam[string](request, "type")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if repo_type != "" {
				opts.Type = repo_type
			}
			sort, err := OptionalParam[string](request, "sort")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if sort != "" {
				opts.Sort = sort
			}
			direction, err := OptionalParam[string](request, "direction")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if direction != "" {
				opts.Direction = direction
			}

			repos, resp, err := client.Repositories.ListByOrg(ctx, org, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list repositories: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != 200 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to list repositories: %s", string(body))), nil
			}

			r, err := json.Marshal(repos)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}
