package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	ghErrors "github.com/github/github-mcp-server/pkg/errors"
	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v73/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetRepositoryRuleset creates a tool to get a specific repository ruleset.
func GetRepositoryRuleset(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_repository_ruleset",
			mcp.WithDescription(t("TOOL_GET_REPOSITORY_RULESET_DESCRIPTION", "Get details of a specific repository ruleset")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_GET_REPOSITORY_RULESET_USER_TITLE", "Get repository ruleset"),
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
			mcp.WithNumber("rulesetId",
				mcp.Required(),
				mcp.Description("Ruleset ID"),
			),
			mcp.WithBoolean("includesParents",
				mcp.Description("Include rulesets configured at higher levels that also apply"),
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
			rulesetID, err := RequiredInt(request, "rulesetId")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			includesParents, err := OptionalParam[bool](request, "includesParents")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			ruleset, resp, err := client.Repositories.GetRuleset(ctx, owner, repo, int64(rulesetID), includesParents)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to get repository ruleset",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			return MarshalledTextResult(ruleset), nil
		}
}

// ListRepositoryRulesets creates a tool to list all repository rulesets.
func ListRepositoryRulesets(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_repository_rulesets",
			mcp.WithDescription(t("TOOL_LIST_REPOSITORY_RULESETS_DESCRIPTION", "List all repository rulesets")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_LIST_REPOSITORY_RULESETS_USER_TITLE", "List repository rulesets"),
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
			mcp.WithBoolean("includesParents",
				mcp.Description("Include rulesets configured at higher levels that also apply"),
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
			includesParents, err := OptionalParam[bool](request, "includesParents")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			opts := &github.RepositoryListRulesetsOptions{
				IncludesParents: &includesParents,
			}

			rulesets, resp, err := client.Repositories.GetAllRulesets(ctx, owner, repo, opts)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to list repository rulesets",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			return MarshalledTextResult(rulesets), nil
		}
}

// GetRepositoryRulesForBranch creates a tool to get all repository rules that apply to a specific branch.
func GetRepositoryRulesForBranch(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_repository_rules_for_branch",
			mcp.WithDescription(t("TOOL_GET_REPOSITORY_RULES_FOR_BRANCH_DESCRIPTION", "Get all repository rules that apply to a specific branch")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_GET_REPOSITORY_RULES_FOR_BRANCH_USER_TITLE", "Get rules for branch"),
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
			mcp.WithString("branch",
				mcp.Required(),
				mcp.Description("Branch name"),
			),
			WithPagination(),
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
			branch, err := RequiredParam[string](request, "branch")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			opts := &github.ListOptions{
				Page:    pagination.Page,
				PerPage: pagination.PerPage,
			}

			branchRules, resp, err := client.Repositories.GetRulesForBranch(ctx, owner, repo, branch, opts)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to get repository rules for branch",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			return MarshalledTextResult(branchRules), nil
		}
}

// GetOrganizationRepositoryRuleset creates a tool to get a specific organization repository ruleset.
func GetOrganizationRepositoryRuleset(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_organization_repository_ruleset",
			mcp.WithDescription(t("TOOL_GET_ORGANIZATION_REPOSITORY_RULESET_DESCRIPTION", "Get details of a specific organization repository ruleset")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_GET_ORGANIZATION_REPOSITORY_RULESET_USER_TITLE", "Get organization repository ruleset"),
				ReadOnlyHint: ToBoolPtr(true),
			}),
			mcp.WithString("org",
				mcp.Required(),
				mcp.Description("Organization name"),
			),
			mcp.WithNumber("rulesetId",
				mcp.Required(),
				mcp.Description("Ruleset ID"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			org, err := RequiredParam[string](request, "org")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			rulesetID, err := RequiredInt(request, "rulesetId")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			ruleset, resp, err := client.Organizations.GetRepositoryRuleset(ctx, org, int64(rulesetID))
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to get organization repository ruleset",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			return MarshalledTextResult(ruleset), nil
		}
}

// ListOrganizationRepositoryRulesets creates a tool to list all organization repository rulesets.
func ListOrganizationRepositoryRulesets(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_organization_repository_rulesets",
			mcp.WithDescription(t("TOOL_LIST_ORGANIZATION_REPOSITORY_RULESETS_DESCRIPTION", "List all organization repository rulesets")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_LIST_ORGANIZATION_REPOSITORY_RULESETS_USER_TITLE", "List organization repository rulesets"),
				ReadOnlyHint: ToBoolPtr(true),
			}),
			mcp.WithString("org",
				mcp.Required(),
				mcp.Description("Organization name"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			org, err := RequiredParam[string](request, "org")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			opts := &github.ListOptions{
				Page:    pagination.Page,
				PerPage: pagination.PerPage,
			}

			rulesets, resp, err := client.Organizations.GetAllRepositoryRulesets(ctx, org, opts)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to list organization repository rulesets",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			return MarshalledTextResult(rulesets), nil
		}
}

// RuleSuite represents a rule suite from GitHub API
type RuleSuite struct {
	ID               *int64           `json:"id,omitempty"`
	ActorID          *int64           `json:"actor_id,omitempty"`
	ActorName        *string          `json:"actor_name,omitempty"`
	BeforeSHA        *string          `json:"before_sha,omitempty"`
	AfterSHA         *string          `json:"after_sha,omitempty"`
	Ref              *string          `json:"ref,omitempty"`
	RepositoryID     *int64           `json:"repository_id,omitempty"`
	RepositoryName   *string          `json:"repository_name,omitempty"`
	PushedAt         *string          `json:"pushed_at,omitempty"`
	Result           *string          `json:"result,omitempty"`
	EvaluationResult *string          `json:"evaluation_result,omitempty"`
	RuleEvaluations  []RuleEvaluation `json:"rule_evaluations,omitempty"`
}

// RuleEvaluation represents a rule evaluation within a rule suite
type RuleEvaluation struct {
	RuleSource  *RuleSource `json:"rule_source,omitempty"`
	Enforcement *string     `json:"enforcement,omitempty"`
	Result      *string     `json:"result,omitempty"`
	RuleType    *string     `json:"rule_type,omitempty"`
	Details     *string     `json:"details,omitempty"`
}

// RuleSource represents the source of a rule
type RuleSource struct {
	Type *string `json:"type,omitempty"`
	ID   *int64  `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// RuleSuitesResponse represents the response from list rule suites API
type RuleSuitesResponse struct {
	RuleSuites []*RuleSuite `json:"rule_suites,omitempty"`
}

// ListRepositoryRuleSuites creates a tool to list rule suites for a repository.
func ListRepositoryRuleSuites(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_repository_rule_suites",
			mcp.WithDescription(t("TOOL_LIST_REPOSITORY_RULE_SUITES_DESCRIPTION", "List rule suites for a repository")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_LIST_REPOSITORY_RULE_SUITES_USER_TITLE", "List repository rule suites"),
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
			mcp.WithString("ref",
				mcp.Description("The name of the ref (branch, tag, etc.) to filter rule suites by"),
			),
			mcp.WithString("timePeriod",
				mcp.Description("The time period to filter by. Options: hour, day, week, month"),
			),
			mcp.WithString("actorName",
				mcp.Description("The handle for the GitHub user account to filter on"),
			),
			mcp.WithString("ruleSuiteResult",
				mcp.Description("The rule suite result to filter by. Options: pass, fail, bypass"),
			),
			WithPagination(),
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

			// Optional parameters
			ref, err := OptionalParam[string](request, "ref")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			timePeriod, err := OptionalParam[string](request, "timePeriod")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			actorName, err := OptionalParam[string](request, "actorName")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			ruleSuiteResult, err := OptionalParam[string](request, "ruleSuiteResult")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			// Build URL with query parameters
			u := fmt.Sprintf("https://api.github.com/repos/%s/%s/rulesets/rule-suites", url.PathEscape(owner), url.PathEscape(repo))
			query := url.Values{}

			if ref != "" {
				query.Add("ref", ref)
			}
			if timePeriod != "" {
				query.Add("time_period", timePeriod)
			}
			if actorName != "" {
				query.Add("actor_name", actorName)
			}
			if ruleSuiteResult != "" {
				query.Add("rule_suite_result", ruleSuiteResult)
			}
			if pagination.Page > 0 {
				query.Add("page", strconv.Itoa(pagination.Page))
			}
			if pagination.PerPage > 0 {
				query.Add("per_page", strconv.Itoa(pagination.PerPage))
			}

			if len(query) > 0 {
				u += "?" + query.Encode()
			}

			req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			req.Header.Set("Accept", "application/vnd.github+json")
			req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

			httpClient := client.Client()
			resp, err := httpClient.Do(req)
			if err != nil {
				var ghResp *github.Response
				if resp != nil {
					ghResp = &github.Response{Response: resp}
				}
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to list repository rule suites",
					ghResp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				var ghResp *github.Response
				if resp != nil {
					ghResp = &github.Response{Response: resp}
				}
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to list repository rule suites",
					ghResp,
					fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)),
				), nil
			}

			var ruleSuites RuleSuitesResponse
			if err := json.Unmarshal(body, &ruleSuites); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}

			return MarshalledTextResult(ruleSuites.RuleSuites), nil
		}
}

// GetRepositoryRuleSuite creates a tool to get details of a specific repository rule suite.
func GetRepositoryRuleSuite(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_repository_rule_suite",
			mcp.WithDescription(t("TOOL_GET_REPOSITORY_RULE_SUITE_DESCRIPTION", "Get details of a specific repository rule suite")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_GET_REPOSITORY_RULE_SUITE_USER_TITLE", "Get repository rule suite"),
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
			mcp.WithNumber("ruleSuiteId",
				mcp.Required(),
				mcp.Description("Rule suite ID"),
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
			ruleSuiteID, err := RequiredInt(request, "ruleSuiteId")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			u := fmt.Sprintf("https://api.github.com/repos/%s/%s/rulesets/rule-suites/%d",
				url.PathEscape(owner), url.PathEscape(repo), ruleSuiteID)

			req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			req.Header.Set("Accept", "application/vnd.github+json")
			req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

			httpClient := client.Client()
			resp, err := httpClient.Do(req)
			if err != nil {
				var ghResp *github.Response
				if resp != nil {
					ghResp = &github.Response{Response: resp}
				}
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to get repository rule suite",
					ghResp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				var ghResp *github.Response
				if resp != nil {
					ghResp = &github.Response{Response: resp}
				}
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to get repository rule suite",
					ghResp,
					fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)),
				), nil
			}

			var ruleSuite RuleSuite
			if err := json.Unmarshal(body, &ruleSuite); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}

			return MarshalledTextResult(ruleSuite), nil
		}
}

// CreateRepositoryRuleset creates a tool to create a new repository ruleset.
func CreateRepositoryRuleset(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("create_repository_ruleset",
			mcp.WithDescription(t("TOOL_CREATE_REPOSITORY_RULESET_DESCRIPTION", "Create a new repository ruleset")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_CREATE_REPOSITORY_RULESET_USER_TITLE", "Create repository ruleset"),
				ReadOnlyHint: ToBoolPtr(false),
			}),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("The name of the ruleset"),
			),
			mcp.WithString("enforcement",
				mcp.Required(),
				mcp.Description("The enforcement level of the ruleset. Can be 'disabled', 'active', or 'evaluate'"),
			),
			mcp.WithString("target",
				mcp.Description("The target of the ruleset. Defaults to 'branch'. Can be one of: 'branch', 'tag', or 'push'"),
			),
			mcp.WithArray("rules",
				mcp.Required(),
				mcp.Description("An array of rules within the ruleset"),
				mcp.Items(
					map[string]any{
						"type": "object",
					},
				),
			),
			mcp.WithObject("conditions",
				mcp.Description("Conditions for when this ruleset applies"),
			),
			mcp.WithArray("bypass_actors",
				mcp.Description("The actors that can bypass the rules in this ruleset"),
				mcp.Items(
					map[string]any{
						"type": "object",
					},
				),
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
			name, err := RequiredParam[string](request, "name")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			enforcement, err := RequiredParam[string](request, "enforcement")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Validate enforcement value
			if enforcement != "disabled" && enforcement != "active" && enforcement != "evaluate" {
				return mcp.NewToolResultError("enforcement must be one of: 'disabled', 'active', 'evaluate'"), nil
			}

			// Parse rules parameter - required array
			rulesObj, ok := request.GetArguments()["rules"].([]interface{})
			if !ok {
				return mcp.NewToolResultError("rules parameter must be an array of rule objects"), nil
			}

			// Optional parameters
			target, err := OptionalParam[string](request, "target")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if target == "" {
				target = "branch" // Default value
			}

			var conditionsObj map[string]interface{}
			if conditionsVal, exists := request.GetArguments()["conditions"]; exists {
				if conditionsMap, ok := conditionsVal.(map[string]interface{}); ok {
					conditionsObj = conditionsMap
				} else {
					return mcp.NewToolResultError("conditions parameter must be an object"), nil
				}
			}

			var bypassActorsObj []interface{}
			if bypassVal, exists := request.GetArguments()["bypass_actors"]; exists {
				if bypassArr, ok := bypassVal.([]interface{}); ok {
					bypassActorsObj = bypassArr
				} else {
					return mcp.NewToolResultError("bypass_actors parameter must be an array of objects"), nil
				}
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			// Build ruleset creation request
			rulesetReq := map[string]any{
				"name":        name,
				"enforcement": enforcement,
				"target":      target,
				"rules":       rulesObj,
			}

			if conditionsObj != nil {
				rulesetReq["conditions"] = conditionsObj
			}
			if bypassActorsObj != nil {
				rulesetReq["bypass_actors"] = bypassActorsObj
			}

			// Convert to JSON for the API request
			jsonData, err := json.Marshal(rulesetReq)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal ruleset request: %w", err)
			}

			// Make the API request
			u := fmt.Sprintf("https://api.github.com/repos/%s/%s/rulesets", url.PathEscape(owner), url.PathEscape(repo))

			// Create a new request with the JSON body
			req, err := http.NewRequestWithContext(ctx, "POST", u, strings.NewReader(string(jsonData)))
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			req.Header.Set("Accept", "application/vnd.github+json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

			// Use the GitHub client's underlying HTTP client to make the request
			httpClient := client.Client()

			resp, err := httpClient.Do(req)
			if err != nil {
				var ghResp *github.Response
				if resp != nil {
					ghResp = &github.Response{Response: resp}
				}
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to create repository ruleset",
					ghResp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}

			if resp.StatusCode != http.StatusCreated {
				var ghResp *github.Response
				if resp != nil {
					ghResp = &github.Response{Response: resp}
				}
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to create repository ruleset",
					ghResp,
					fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)),
				), nil
			}

			var createdRuleset map[string]any
			if err := json.Unmarshal(body, &createdRuleset); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}

			return MarshalledTextResult(createdRuleset), nil
		}
}

// CreateOrganizationRepositoryRuleset creates a tool to create a new organization repository ruleset.
func CreateOrganizationRepositoryRuleset(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("create_organization_repository_ruleset",
			mcp.WithDescription(t("TOOL_CREATE_ORGANIZATION_REPOSITORY_RULESET_DESCRIPTION", "Create a new organization repository ruleset")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_CREATE_ORGANIZATION_REPOSITORY_RULESET_USER_TITLE", "Create organization repository ruleset"),
				ReadOnlyHint: ToBoolPtr(false),
			}),
			mcp.WithString("org",
				mcp.Required(),
				mcp.Description("Organization name"),
			),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("The name of the ruleset"),
			),
			mcp.WithString("enforcement",
				mcp.Required(),
				mcp.Description("The enforcement level of the ruleset. Can be 'disabled', 'active', or 'evaluate'"),
			),
			mcp.WithString("target",
				mcp.Description("The target of the ruleset. Defaults to 'branch'. Can be one of: 'branch', 'tag', or 'push'"),
			),
			mcp.WithArray("rules",
				mcp.Required(),
				mcp.Description("An array of rules within the ruleset"),
				mcp.Items(
					map[string]any{
						"type": "object",
					},
				),
			),
			mcp.WithObject("conditions",
				mcp.Description("Conditions for when this ruleset applies"),
			),
			mcp.WithArray("bypass_actors",
				mcp.Description("The actors that can bypass the rules in this ruleset"),
				mcp.Items(
					map[string]any{
						"type": "object",
					},
				),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			org, err := RequiredParam[string](request, "org")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			name, err := RequiredParam[string](request, "name")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			enforcement, err := RequiredParam[string](request, "enforcement")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Validate enforcement value
			if enforcement != "disabled" && enforcement != "active" && enforcement != "evaluate" {
				return mcp.NewToolResultError("enforcement must be one of: 'disabled', 'active', 'evaluate'"), nil
			}

			// Parse rules parameter - required array
			rulesObj, ok := request.GetArguments()["rules"].([]interface{})
			if !ok {
				return mcp.NewToolResultError("rules parameter must be an array of rule objects"), nil
			}

			// Optional parameters
			target, err := OptionalParam[string](request, "target")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if target == "" {
				target = "branch" // Default value
			}

			var conditionsObj map[string]interface{}
			if conditionsVal, exists := request.GetArguments()["conditions"]; exists {
				if conditionsMap, ok := conditionsVal.(map[string]interface{}); ok {
					conditionsObj = conditionsMap
				} else {
					return mcp.NewToolResultError("conditions parameter must be an object"), nil
				}
			}

			var bypassActorsObj []interface{}
			if bypassVal, exists := request.GetArguments()["bypass_actors"]; exists {
				if bypassArr, ok := bypassVal.([]interface{}); ok {
					bypassActorsObj = bypassArr
				} else {
					return mcp.NewToolResultError("bypass_actors parameter must be an array of objects"), nil
				}
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			// Build ruleset creation request
			rulesetReq := map[string]any{
				"name":        name,
				"enforcement": enforcement,
				"target":      target,
				"rules":       rulesObj,
			}

			if conditionsObj != nil {
				rulesetReq["conditions"] = conditionsObj
			}
			if bypassActorsObj != nil {
				rulesetReq["bypass_actors"] = bypassActorsObj
			}

			// Convert to JSON for the API request
			jsonData, err := json.Marshal(rulesetReq)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal ruleset request: %w", err)
			}

			// Make the API request
			u := fmt.Sprintf("https://api.github.com/orgs/%s/rulesets", url.PathEscape(org))

			// Create a new request with the JSON body
			req, err := http.NewRequestWithContext(ctx, "POST", u, strings.NewReader(string(jsonData)))
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			req.Header.Set("Accept", "application/vnd.github+json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

			// Use the GitHub client's underlying HTTP client to make the request
			httpClient := client.Client()

			resp, err := httpClient.Do(req)
			if err != nil {
				var ghResp *github.Response
				if resp != nil {
					ghResp = &github.Response{Response: resp}
				}
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to create organization repository ruleset",
					ghResp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}

			if resp.StatusCode != http.StatusCreated {
				var ghResp *github.Response
				if resp != nil {
					ghResp = &github.Response{Response: resp}
				}
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to create organization repository ruleset",
					ghResp,
					fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)),
				), nil
			}

			var createdRuleset map[string]any
			if err := json.Unmarshal(body, &createdRuleset); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}

			return MarshalledTextResult(createdRuleset), nil
		}
}

// CreateEnterpriseRepositoryRuleset creates a tool to create a new enterprise repository ruleset.
func CreateEnterpriseRepositoryRuleset(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("create_enterprise_repository_ruleset",
			mcp.WithDescription(t("TOOL_CREATE_ENTERPRISE_REPOSITORY_RULESET_DESCRIPTION", "Create a new enterprise repository ruleset")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_CREATE_ENTERPRISE_REPOSITORY_RULESET_USER_TITLE", "Create enterprise repository ruleset"),
				ReadOnlyHint: ToBoolPtr(false),
			}),
			mcp.WithString("enterprise",
				mcp.Required(),
				mcp.Description("Enterprise name"),
			),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("The name of the ruleset"),
			),
			mcp.WithString("enforcement",
				mcp.Required(),
				mcp.Description("The enforcement level of the ruleset. Can be 'disabled', 'active', or 'evaluate'"),
			),
			mcp.WithString("target",
				mcp.Description("The target of the ruleset. Defaults to 'branch'. Can be one of: 'branch', 'tag', or 'push'"),
			),
			mcp.WithArray("rules",
				mcp.Required(),
				mcp.Description("An array of rules within the ruleset"),
				mcp.Items(
					map[string]any{
						"type": "object",
					},
				),
			),
			mcp.WithObject("conditions",
				mcp.Description("Conditions for when this ruleset applies"),
			),
			mcp.WithArray("bypass_actors",
				mcp.Description("The actors that can bypass the rules in this ruleset"),
				mcp.Items(
					map[string]any{
						"type": "object",
					},
				),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			enterprise, err := RequiredParam[string](request, "enterprise")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			name, err := RequiredParam[string](request, "name")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			enforcement, err := RequiredParam[string](request, "enforcement")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Validate enforcement value
			if enforcement != "disabled" && enforcement != "active" && enforcement != "evaluate" {
				return mcp.NewToolResultError("enforcement must be one of: 'disabled', 'active', 'evaluate'"), nil
			}

			// Parse rules parameter - required array
			rulesObj, ok := request.GetArguments()["rules"].([]interface{})
			if !ok {
				return mcp.NewToolResultError("rules parameter must be an array of rule objects"), nil
			}

			// Optional parameters
			target, err := OptionalParam[string](request, "target")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if target == "" {
				target = "branch" // Default value
			}

			var conditionsObj map[string]interface{}
			if conditionsVal, exists := request.GetArguments()["conditions"]; exists {
				if conditionsMap, ok := conditionsVal.(map[string]interface{}); ok {
					conditionsObj = conditionsMap
				} else {
					return mcp.NewToolResultError("conditions parameter must be an object"), nil
				}
			}

			var bypassActorsObj []interface{}
			if bypassVal, exists := request.GetArguments()["bypass_actors"]; exists {
				if bypassArr, ok := bypassVal.([]interface{}); ok {
					bypassActorsObj = bypassArr
				} else {
					return mcp.NewToolResultError("bypass_actors parameter must be an array of objects"), nil
				}
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			// Build ruleset creation request
			rulesetReq := map[string]any{
				"name":        name,
				"enforcement": enforcement,
				"target":      target,
				"rules":       rulesObj,
			}

			if conditionsObj != nil {
				rulesetReq["conditions"] = conditionsObj
			}
			if bypassActorsObj != nil {
				rulesetReq["bypass_actors"] = bypassActorsObj
			}

			// Convert to JSON for the API request
			jsonData, err := json.Marshal(rulesetReq)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal ruleset request: %w", err)
			}

			// Make the API request
			u := fmt.Sprintf("https://api.github.com/enterprises/%s/rulesets", url.PathEscape(enterprise))

			// Create a new request with the JSON body
			req, err := http.NewRequestWithContext(ctx, "POST", u, strings.NewReader(string(jsonData)))
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			req.Header.Set("Accept", "application/vnd.github+json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

			// Use the GitHub client's underlying HTTP client to make the request
			httpClient := client.Client()

			resp, err := httpClient.Do(req)
			if err != nil {
				var ghResp *github.Response
				if resp != nil {
					ghResp = &github.Response{Response: resp}
				}
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to create enterprise repository ruleset",
					ghResp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}

			if resp.StatusCode != http.StatusCreated {
				var ghResp *github.Response
				if resp != nil {
					ghResp = &github.Response{Response: resp}
				}
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to create enterprise repository ruleset",
					ghResp,
					fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)),
				), nil
			}

			var createdRuleset map[string]any
			if err := json.Unmarshal(body, &createdRuleset); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}

			return MarshalledTextResult(createdRuleset), nil
		}
}
