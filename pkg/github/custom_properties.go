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

func GetRepositoryCustomProperties(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_repository_custom_properties",
			mcp.WithDescription(t("TOOL_GET_REPOSITORY_CUSTOM_PROPERTIES_DESCRIPTION", "Get custom properties for a repository")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				ReadOnlyHint: ToBoolPtr(true),
			}),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := RequiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := RequiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			req, err := client.NewRequest("GET", fmt.Sprintf("repos/%s/%s/properties/values", owner, repo), nil)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			var props []*github.CustomProperty
			_, err = client.Do(ctx, req, &props)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return MarshalledTextResult(props)
		}
}

func CreateOrUpdateRepositoryCustomProperties(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool(
			"create_or_update_repository_custom_properties",
			mcp.WithDescription(t("GITHUB_CREATE_OR_UPDATE_REPOSITORY_CUSTOM_PROPERTIES_DESCRIPTION", "Create or update repository custom properties")),
			mcp.WithString("owner", mcp.Required(), mcp.Description("Repository owner")),
			mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
			mcp.WithString("properties", mcp.Required(), mcp.Description("Custom properties as JSON array")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := RequiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := RequiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			propertiesStr, err := RequiredParam[string](request, "properties")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var props []*github.CustomProperty
			if err := json.Unmarshal([]byte(propertiesStr), &props); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			client, err := getClient(ctx)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			req, err := client.NewRequest("PATCH", fmt.Sprintf("repos/%s/%s/properties/values", owner, repo), map[string]interface{}{"properties": props})
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			_, err = client.Do(ctx, req, nil)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText("Custom properties updated successfully"), nil
		}
}

func GetOrganizationCustomProperties(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool(
			"get_organization_custom_properties",
			mcp.WithDescription(t("GITHUB_GET_ORGANIZATION_CUSTOM_PROPERTIES_DESCRIPTION", "Get organization custom properties")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				ReadOnlyHint: ToBoolPtr(true),
			}),
			mcp.WithString("org", mcp.Required(), mcp.Description("Organization name")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			org, err := RequiredParam[string](request, "org")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			client, err := getClient(ctx)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			req, err := client.NewRequest("GET", fmt.Sprintf("orgs/%s/properties/schema", org), nil)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var props []*github.CustomProperty
			_, err = client.Do(ctx, req, &props)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return MarshalledTextResult("Organization custom properties for "+org, props)
		}
}

func CreateOrUpdateOrganizationCustomProperties(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool(
			"create_or_update_organization_custom_properties",
			mcp.WithDescription(t("GITHUB_CREATE_OR_UPDATE_ORGANIZATION_CUSTOM_PROPERTIES_DESCRIPTION", "Create or update organization custom properties")),
			mcp.WithString("org", mcp.Required(), mcp.Description("Organization name")),
			mcp.WithString("properties", mcp.Required(), mcp.Description("Custom properties as JSON array")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			org, err := RequiredParam[string](request, "org")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			propertiesStr, err := RequiredParam[string](request, "properties")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var props []*github.CustomProperty
			if err := json.Unmarshal([]byte(propertiesStr), &props); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			client, err := getClient(ctx)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			req, err := client.NewRequest("PATCH", fmt.Sprintf("orgs/%s/properties/schema", org), map[string]interface{}{"properties": props})
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			_, err = client.Do(ctx, req, nil)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText("Custom properties updated successfully"), nil
		}
}

func GetEnterpriseCustomProperties(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool(
			"get_enterprise_custom_properties",
			mcp.WithDescription(t("GITHUB_GET_ENTERPRISE_CUSTOM_PROPERTIES_DESCRIPTION", "Get enterprise custom properties")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				ReadOnlyHint: ToBoolPtr(true),
			}),
			mcp.WithString("enterprise", mcp.Required(), mcp.Description("Enterprise name")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			enterprise, err := RequiredParam[string](request, "enterprise")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			client, err := getClient(ctx)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			req, err := client.NewRequest("GET", fmt.Sprintf("enterprises/%s/properties/schema", enterprise), nil)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var props []*github.CustomProperty
			_, err = client.Do(ctx, req, &props)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Enterprise custom properties for %s: %v", enterprise, props)), nil
		}
}

func CreateOrUpdateEnterpriseCustomProperties(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool(
			"create_or_update_enterprise_custom_properties",
			mcp.WithDescription(t("GITHUB_CREATE_OR_UPDATE_ENTERPRISE_CUSTOM_PROPERTIES_DESCRIPTION", "Create or update enterprise custom properties")),
			mcp.WithString("enterprise", mcp.Required(), mcp.Description("Enterprise name")),
			mcp.WithString("properties", mcp.Required(), mcp.Description("Custom properties as JSON array")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			enterprise, err := RequiredParam[string](request, "enterprise")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			propertiesStr, err := RequiredParam[string](request, "properties")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var props []*github.CustomProperty
			if err := json.Unmarshal([]byte(propertiesStr), &props); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			client, err := getClient(ctx)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			req, err := client.NewRequest("PATCH", fmt.Sprintf("enterprises/%s/properties/schema", enterprise), map[string]interface{}{"properties": props})
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			_, err = client.Do(ctx, req, nil)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText("Custom properties updated successfully"), nil
		}
}
