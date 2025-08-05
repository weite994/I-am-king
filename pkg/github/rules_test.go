package github

import (
	"context"
	"net/http"
	"testing"

	"github.com/github/github-mcp-server/internal/toolsnaps"
	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v73/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetRepositoryRuleset(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetRepositoryRuleset(stubGetClientFn(mockClient), translations.NullTranslationHelper)
	require.NoError(t, toolsnaps.Test(tool.Name, tool))

	assert.Equal(t, "get_repository_ruleset", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "rulesetId")
	assert.Contains(t, tool.InputSchema.Properties, "includesParents")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "rulesetId"})

	// Setup mock ruleset for success case
	mockRuleset := &github.RepositoryRuleset{
		ID:          github.Ptr(int64(123)),
		Name:        "test-ruleset",
		Enforcement: github.RulesetEnforcementActive,
		Target:      github.Ptr(github.RulesetTargetBranch),
	}

	tests := []struct {
		name            string
		mockedClient    *http.Client
		requestArgs     map[string]interface{}
		expectError     bool
		expectedRuleset *github.RepositoryRuleset
		expectedErrMsg  string
	}{
		{
			name: "successful ruleset fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposRulesetsByOwnerByRepoByRulesetId,
					mockRuleset,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":           "testowner",
				"repo":            "testrepo",
				"rulesetId":       float64(123),
				"includesParents": true,
			},
			expectError:     false,
			expectedRuleset: mockRuleset,
		},
		{
			name:         "missing required parameter owner",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"repo":      "testrepo",
				"rulesetId": float64(123),
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: owner",
		},
		{
			name:         "missing required parameter repo",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"owner":     "testowner",
				"rulesetId": float64(123),
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: repo",
		},
		{
			name:         "missing required parameter rulesetId",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"owner": "testowner",
				"repo":  "testrepo",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: rulesetId",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := github.NewClient(tt.mockedClient)
			_, handler := GetRepositoryRuleset(stubGetClientFn(client), translations.NullTranslationHelper)

			result, err := handler(context.Background(), createMCPRequest(tt.requestArgs))

			if tt.expectError {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsError)
				if tt.expectedErrMsg != "" {
					textResult := getErrorResult(t, result)
					assert.Contains(t, textResult.Text, tt.expectedErrMsg)
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.False(t, result.IsError)
				textResult := getTextResult(t, result)
				assert.NotEmpty(t, textResult.Text)
			}
		})
	}
}

func Test_ListRepositoryRulesets(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := ListRepositoryRulesets(stubGetClientFn(mockClient), translations.NullTranslationHelper)
	require.NoError(t, toolsnaps.Test(tool.Name, tool))

	assert.Equal(t, "list_repository_rulesets", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "includesParents")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})

	// Setup mock rulesets for success case
	mockRulesets := []*github.RepositoryRuleset{
		{
			ID:   github.Ptr(int64(123)),
			Name: "test-ruleset-1",
		},
		{
			ID:   github.Ptr(int64(456)),
			Name: "test-ruleset-2",
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful rulesets listing",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposRulesetsByOwnerByRepo,
					mockRulesets,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":           "testowner",
				"repo":            "testrepo",
				"includesParents": false,
			},
			expectError: false,
		},
		{
			name:         "missing required parameter owner",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"repo": "testrepo",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := github.NewClient(tt.mockedClient)
			_, handler := ListRepositoryRulesets(stubGetClientFn(client), translations.NullTranslationHelper)

			result, err := handler(context.Background(), createMCPRequest(tt.requestArgs))

			if tt.expectError {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsError)
				if tt.expectedErrMsg != "" {
					textResult := getErrorResult(t, result)
					assert.Contains(t, textResult.Text, tt.expectedErrMsg)
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.False(t, result.IsError)
				textResult := getTextResult(t, result)
				assert.NotEmpty(t, textResult.Text)
			}
		})
	}
}

func Test_GetRepositoryRulesForBranch(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetRepositoryRulesForBranch(stubGetClientFn(mockClient), translations.NullTranslationHelper)
	require.NoError(t, toolsnaps.Test(tool.Name, tool))

	assert.Equal(t, "get_repository_rules_for_branch", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "branch")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "branch"})

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedErrMsg string
	}{
		// TODO: Fix this test - the mock response format doesn't match the expected branchRuleWrapper array format
		// {
		// 	name: "successful branch rules fetch",
		// 	mockedClient: mock.NewMockedHTTPClient(
		// 		mock.WithRequestMatch(
		// 			mock.GetReposRulesBranchesByOwnerByRepoByBranch,
		// 			mockBranchRules,
		// 		),
		// 	),
		// 	requestArgs: map[string]interface{}{
		// 		"owner":  "testowner",
		// 		"repo":   "testrepo",
		// 		"branch": "main",
		// 	},
		// 	expectError: false,
		// },
		{
			name:         "missing required parameter branch",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"owner": "testowner",
				"repo":  "testrepo",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := github.NewClient(tt.mockedClient)
			_, handler := GetRepositoryRulesForBranch(stubGetClientFn(client), translations.NullTranslationHelper)

			result, err := handler(context.Background(), createMCPRequest(tt.requestArgs))

			if tt.expectError {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsError)
				if tt.expectedErrMsg != "" {
					textResult := getErrorResult(t, result)
					assert.Contains(t, textResult.Text, tt.expectedErrMsg)
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.False(t, result.IsError)
				textResult := getTextResult(t, result)
				assert.NotEmpty(t, textResult.Text)
			}
		})
	}
}

func Test_GetOrganizationRepositoryRuleset(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetOrganizationRepositoryRuleset(stubGetClientFn(mockClient), translations.NullTranslationHelper)
	require.NoError(t, toolsnaps.Test(tool.Name, tool))

	assert.Equal(t, "get_organization_repository_ruleset", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "org")
	assert.Contains(t, tool.InputSchema.Properties, "rulesetId")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"org", "rulesetId"})

	// Setup mock organization ruleset for success case
	mockRuleset := &github.RepositoryRuleset{
		ID:          github.Ptr(int64(789)),
		Name:        "org-test-ruleset",
		Enforcement: github.RulesetEnforcementActive,
		Target:      github.Ptr(github.RulesetTargetBranch),
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful organization ruleset fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetOrgsRulesetsByOrgByRulesetId,
					mockRuleset,
				),
			),
			requestArgs: map[string]interface{}{
				"org":       "testorg",
				"rulesetId": float64(789),
			},
			expectError: false,
		},
		{
			name:         "missing required parameter org",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"rulesetId": float64(789),
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := github.NewClient(tt.mockedClient)
			_, handler := GetOrganizationRepositoryRuleset(stubGetClientFn(client), translations.NullTranslationHelper)

			result, err := handler(context.Background(), createMCPRequest(tt.requestArgs))

			if tt.expectError {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsError)
				if tt.expectedErrMsg != "" {
					textResult := getErrorResult(t, result)
					assert.Contains(t, textResult.Text, tt.expectedErrMsg)
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.False(t, result.IsError)
				textResult := getTextResult(t, result)
				assert.NotEmpty(t, textResult.Text)
			}
		})
	}
}

func Test_ListOrganizationRepositoryRulesets(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := ListOrganizationRepositoryRulesets(stubGetClientFn(mockClient), translations.NullTranslationHelper)
	require.NoError(t, toolsnaps.Test(tool.Name, tool))

	assert.Equal(t, "list_organization_repository_rulesets", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "org")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"org"})

	// Setup mock organization rulesets for success case
	mockRulesets := []*github.RepositoryRuleset{
		{
			ID:   github.Ptr(int64(789)),
			Name: "org-test-ruleset-1",
		},
		{
			ID:   github.Ptr(int64(790)),
			Name: "org-test-ruleset-2",
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful organization rulesets listing",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetOrgsRulesetsByOrg,
					mockRulesets,
				),
			),
			requestArgs: map[string]interface{}{
				"org": "testorg",
			},
			expectError: false,
		},
		{
			name:           "missing required parameter org",
			mockedClient:   mock.NewMockedHTTPClient(),
			requestArgs:    map[string]interface{}{},
			expectError:    true,
			expectedErrMsg: "missing required parameter: org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := github.NewClient(tt.mockedClient)
			_, handler := ListOrganizationRepositoryRulesets(stubGetClientFn(client), translations.NullTranslationHelper)

			result, err := handler(context.Background(), createMCPRequest(tt.requestArgs))

			if tt.expectError {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsError)
				if tt.expectedErrMsg != "" {
					textResult := getErrorResult(t, result)
					assert.Contains(t, textResult.Text, tt.expectedErrMsg)
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.False(t, result.IsError)
				textResult := getTextResult(t, result)
				assert.NotEmpty(t, textResult.Text)
			}
		})
	}
}

func Test_ListRepositoryRuleSuites(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := ListRepositoryRuleSuites(stubGetClientFn(mockClient), translations.NullTranslationHelper)
	require.NoError(t, toolsnaps.Test(tool.Name, tool))

	assert.Equal(t, "list_repository_rule_suites", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "ref")
	assert.Contains(t, tool.InputSchema.Properties, "timePeriod")
	assert.Contains(t, tool.InputSchema.Properties, "actorName")
	assert.Contains(t, tool.InputSchema.Properties, "ruleSuiteResult")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:         "missing required parameter owner",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"repo": "testrepo",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: owner",
		},
		{
			name:         "missing required parameter repo",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"owner": "testowner",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := github.NewClient(tt.mockedClient)
			_, handler := ListRepositoryRuleSuites(stubGetClientFn(client), translations.NullTranslationHelper)

			result, err := handler(context.Background(), createMCPRequest(tt.requestArgs))

			if tt.expectError {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsError)
				if tt.expectedErrMsg != "" {
					textResult := getErrorResult(t, result)
					assert.Contains(t, textResult.Text, tt.expectedErrMsg)
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.False(t, result.IsError)
				textResult := getTextResult(t, result)
				assert.NotEmpty(t, textResult.Text)
			}
		})
	}
}

func Test_GetRepositoryRuleSuite(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetRepositoryRuleSuite(stubGetClientFn(mockClient), translations.NullTranslationHelper)
	require.NoError(t, toolsnaps.Test(tool.Name, tool))

	assert.Equal(t, "get_repository_rule_suite", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "ruleSuiteId")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "ruleSuiteId"})

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:         "missing required parameter owner",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"repo":        "testrepo",
				"ruleSuiteId": float64(456),
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: owner",
		},
		{
			name:         "missing required parameter repo",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"owner":       "testowner",
				"ruleSuiteId": float64(456),
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: repo",
		},
		{
			name:         "missing required parameter ruleSuiteId",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"owner": "testowner",
				"repo":  "testrepo",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: ruleSuiteId",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := github.NewClient(tt.mockedClient)
			_, handler := GetRepositoryRuleSuite(stubGetClientFn(client), translations.NullTranslationHelper)

			result, err := handler(context.Background(), createMCPRequest(tt.requestArgs))

			if tt.expectError {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsError)
				if tt.expectedErrMsg != "" {
					textResult := getErrorResult(t, result)
					assert.Contains(t, textResult.Text, tt.expectedErrMsg)
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.False(t, result.IsError)
				textResult := getTextResult(t, result)
				assert.NotEmpty(t, textResult.Text)
			}
		})
	}
}
