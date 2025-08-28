package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v74/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ManageGist creates a consolidated tool to perform CRUD operations on gists
func ManageGist(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("manage_gist",
			mcp.WithDescription(t("TOOL_MANAGE_GIST_DESCRIPTION", "Manage gists with various operations: list, create, update, get")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_MANAGE_GIST", "Manage Gist"),
				ReadOnlyHint: ToBoolPtr(false),
			}),
			mcp.WithString("operation",
				mcp.Required(),
				mcp.Description("Operation to perform: 'list', 'create', 'update', 'get'"),
				mcp.Enum("list", "create", "update", "get"),
			),
			// Parameters for list operation
			mcp.WithString("username",
				mcp.Description("GitHub username (omit for authenticated user's gists, used for 'list' and 'get' operations)"),
			),
			mcp.WithString("since",
				mcp.Description("Only gists updated after this time (ISO 8601 timestamp, used for 'list' operation)"),
			),
			// Parameters for create/update operations
			mcp.WithString("gist_id",
				mcp.Description("ID of the gist (required for 'update' and 'get' operations)"),
			),
			mcp.WithString("description",
				mcp.Description("Description of the gist (used for 'create' and 'update' operations)"),
			),
			mcp.WithString("filename",
				mcp.Description("Filename for gist file (required for 'create' and 'update' operations)"),
			),
			mcp.WithString("content",
				mcp.Description("Content for gist file (required for 'create' and 'update' operations)"),
			),
			mcp.WithBoolean("public",
				mcp.Description("Whether the gist is public (used for 'create' operation)"),
				mcp.DefaultBool(false),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			operation, err := RequiredParam[string](request, "operation")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			switch operation {
			case "list":
				return handleListGists(ctx, getClient, request)
			case "create":
				return handleCreateGist(ctx, getClient, request)
			case "update":
				return handleUpdateGist(ctx, getClient, request)
			case "get":
				return handleGetGist(ctx, getClient, request)
			default:
				return mcp.NewToolResultError(fmt.Sprintf("unsupported operation: %s", operation)), nil
			}
		}
}

func handleListGists(ctx context.Context, getClient GetClientFn, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username, err := OptionalParam[string](request, "username")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	since, err := OptionalParam[string](request, "since")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pagination, err := OptionalPaginationParams(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	opts := &github.GistListOptions{
		ListOptions: github.ListOptions{
			Page:    pagination.Page,
			PerPage: pagination.PerPage,
		},
	}

	// Parse since timestamp if provided
	if since != "" {
		sinceTime, err := parseISOTimestamp(since)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid since timestamp: %v", err)), nil
		}
		opts.Since = sinceTime
	}

	client, err := getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub client: %w", err)
	}

	gists, resp, err := client.Gists.List(ctx, username, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list gists: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to list gists: %s", string(body))), nil
	}

	r, err := json.Marshal(gists)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return mcp.NewToolResultText(string(r)), nil
}

func handleCreateGist(ctx context.Context, getClient GetClientFn, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	description, err := OptionalParam[string](request, "description")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	filename, err := RequiredParam[string](request, "filename")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	content, err := RequiredParam[string](request, "content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	public, err := OptionalParam[bool](request, "public")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	files := make(map[github.GistFilename]github.GistFile)
	files[github.GistFilename(filename)] = github.GistFile{
		Filename: github.Ptr(filename),
		Content:  github.Ptr(content),
	}

	gist := &github.Gist{
		Files:       files,
		Public:      github.Ptr(public),
		Description: github.Ptr(description),
	}

	client, err := getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub client: %w", err)
	}

	createdGist, resp, err := client.Gists.Create(ctx, gist)
	if err != nil {
		return nil, fmt.Errorf("failed to create gist: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to create gist: %s", string(body))), nil
	}

	r, err := json.Marshal(createdGist)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return mcp.NewToolResultText(string(r)), nil
}

func handleUpdateGist(ctx context.Context, getClient GetClientFn, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gistID, err := RequiredParam[string](request, "gist_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	description, err := OptionalParam[string](request, "description")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	filename, err := RequiredParam[string](request, "filename")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	content, err := RequiredParam[string](request, "content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	files := make(map[github.GistFilename]github.GistFile)
	files[github.GistFilename(filename)] = github.GistFile{
		Filename: github.Ptr(filename),
		Content:  github.Ptr(content),
	}

	gist := &github.Gist{
		Files:       files,
		Description: github.Ptr(description),
	}

	client, err := getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub client: %w", err)
	}

	updatedGist, resp, err := client.Gists.Edit(ctx, gistID, gist)
	if err != nil {
		return nil, fmt.Errorf("failed to update gist: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to update gist: %s", string(body))), nil
	}

	r, err := json.Marshal(updatedGist)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return mcp.NewToolResultText(string(r)), nil
}

func handleGetGist(ctx context.Context, getClient GetClientFn, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gistID, err := RequiredParam[string](request, "gist_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub client: %w", err)
	}

	gist, resp, err := client.Gists.Get(ctx, gistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get gist: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to get gist: %s", string(body))), nil
	}

	r, err := json.Marshal(gist)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return mcp.NewToolResultText(string(r)), nil
}

// ListGists creates a tool to list gists for a user
func ListGists(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_gists",
			mcp.WithDescription(t("TOOL_LIST_GISTS_DESCRIPTION", "List gists for a user")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_LIST_GISTS", "List Gists"),
				ReadOnlyHint: ToBoolPtr(true),
			}),
			mcp.WithString("username",
				mcp.Description("GitHub username (omit for authenticated user's gists)"),
			),
			mcp.WithString("since",
				mcp.Description("Only gists updated after this time (ISO 8601 timestamp)"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			username, err := OptionalParam[string](request, "username")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			since, err := OptionalParam[string](request, "since")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.GistListOptions{
				ListOptions: github.ListOptions{
					Page:    pagination.Page,
					PerPage: pagination.PerPage,
				},
			}

			// Parse since timestamp if provided
			if since != "" {
				sinceTime, err := parseISOTimestamp(since)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("invalid since timestamp: %v", err)), nil
				}
				opts.Since = sinceTime
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			gists, resp, err := client.Gists.List(ctx, username, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list gists: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to list gists: %s", string(body))), nil
			}

			r, err := json.Marshal(gists)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// CreateGist creates a tool to create a new gist
func CreateGist(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("create_gist",
			mcp.WithDescription(t("TOOL_CREATE_GIST_DESCRIPTION", "Create a new gist")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_CREATE_GIST", "Create Gist"),
				ReadOnlyHint: ToBoolPtr(false),
			}),
			mcp.WithString("description",
				mcp.Description("Description of the gist"),
			),
			mcp.WithString("filename",
				mcp.Required(),
				mcp.Description("Filename for simple single-file gist creation"),
			),
			mcp.WithString("content",
				mcp.Required(),
				mcp.Description("Content for simple single-file gist creation"),
			),
			mcp.WithBoolean("public",
				mcp.Description("Whether the gist is public"),
				mcp.DefaultBool(false),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			description, err := OptionalParam[string](request, "description")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			filename, err := RequiredParam[string](request, "filename")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			content, err := RequiredParam[string](request, "content")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			public, err := OptionalParam[bool](request, "public")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			files := make(map[github.GistFilename]github.GistFile)
			files[github.GistFilename(filename)] = github.GistFile{
				Filename: github.Ptr(filename),
				Content:  github.Ptr(content),
			}

			gist := &github.Gist{
				Files:       files,
				Public:      github.Ptr(public),
				Description: github.Ptr(description),
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			createdGist, resp, err := client.Gists.Create(ctx, gist)
			if err != nil {
				return nil, fmt.Errorf("failed to create gist: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to create gist: %s", string(body))), nil
			}

			r, err := json.Marshal(createdGist)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// UpdateGist creates a tool to edit an existing gist
func UpdateGist(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("update_gist",
			mcp.WithDescription(t("TOOL_UPDATE_GIST_DESCRIPTION", "Update an existing gist")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_UPDATE_GIST", "Update Gist"),
				ReadOnlyHint: ToBoolPtr(false),
			}),
			mcp.WithString("gist_id",
				mcp.Required(),
				mcp.Description("ID of the gist to update"),
			),
			mcp.WithString("description",
				mcp.Description("Updated description of the gist"),
			),
			mcp.WithString("filename",
				mcp.Required(),
				mcp.Description("Filename to update or create"),
			),
			mcp.WithString("content",
				mcp.Required(),
				mcp.Description("Content for the file"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			gistID, err := RequiredParam[string](request, "gist_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			description, err := OptionalParam[string](request, "description")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			filename, err := RequiredParam[string](request, "filename")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			content, err := RequiredParam[string](request, "content")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			files := make(map[github.GistFilename]github.GistFile)
			files[github.GistFilename(filename)] = github.GistFile{
				Filename: github.Ptr(filename),
				Content:  github.Ptr(content),
			}

			gist := &github.Gist{
				Files:       files,
				Description: github.Ptr(description),
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			updatedGist, resp, err := client.Gists.Edit(ctx, gistID, gist)
			if err != nil {
				return nil, fmt.Errorf("failed to update gist: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to update gist: %s", string(body))), nil
			}

			r, err := json.Marshal(updatedGist)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}
