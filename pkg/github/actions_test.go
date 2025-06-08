package github

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v72/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ListWorkflows(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := ListWorkflows(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "list_workflows", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "per_page")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]any
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful workflow listing",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposActionsWorkflowsByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						workflows := &github.Workflows{
							TotalCount: github.Int(2),
							Workflows: []*github.Workflow{
								{
									ID:        github.Int64(123),
									Name:      github.String("CI"),
									Path:      github.String(".github/workflows/ci.yml"),
									State:     github.String("active"),
									CreatedAt: &github.Timestamp{},
									UpdatedAt: &github.Timestamp{},
									URL:       github.String("https://api.github.com/repos/owner/repo/actions/workflows/123"),
									HTMLURL:   github.String("https://github.com/owner/repo/actions/workflows/ci.yml"),
									BadgeURL:  github.String("https://github.com/owner/repo/workflows/CI/badge.svg"),
									NodeID:    github.String("W_123"),
								},
								{
									ID:        github.Int64(456),
									Name:      github.String("Deploy"),
									Path:      github.String(".github/workflows/deploy.yml"),
									State:     github.String("active"),
									CreatedAt: &github.Timestamp{},
									UpdatedAt: &github.Timestamp{},
									URL:       github.String("https://api.github.com/repos/owner/repo/actions/workflows/456"),
									HTMLURL:   github.String("https://github.com/owner/repo/actions/workflows/deploy.yml"),
									BadgeURL:  github.String("https://github.com/owner/repo/workflows/Deploy/badge.svg"),
									NodeID:    github.String("W_456"),
								},
							},
						}
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(workflows)
					}),
				),
			),
			requestArgs: map[string]any{
				"owner": "owner",
				"repo":  "repo",
			},
			expectError: false,
		},
		{
			name:         "missing required parameter owner",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]any{
				"repo": "repo",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: owner",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := ListWorkflows(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			require.NoError(t, err)
			require.Equal(t, tc.expectError, result.IsError)

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			if tc.expectedErrMsg != "" {
				assert.Equal(t, tc.expectedErrMsg, textContent.Text)
				return
			}

			// Unmarshal and verify the result
			var response github.Workflows
			err = json.Unmarshal([]byte(textContent.Text), &response)
			require.NoError(t, err)
			assert.NotNil(t, response.TotalCount)
			assert.Greater(t, *response.TotalCount, 0)
			assert.NotEmpty(t, response.Workflows)
		})
	}
}

func Test_RunWorkflow(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := RunWorkflow(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "run_workflow", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "workflow_file")
	assert.Contains(t, tool.InputSchema.Properties, "ref")
	assert.Contains(t, tool.InputSchema.Properties, "inputs")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "workflow_file", "ref"})

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]any
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful workflow run",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposActionsWorkflowsDispatchesByOwnerByRepoByWorkflowId,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNoContent)
					}),
				),
			),
			requestArgs: map[string]any{
				"owner":         "owner",
				"repo":          "repo",
				"workflow_file": "ci.yml",
				"ref":           "main",
			},
			expectError: false,
		},
		{
			name:         "missing required parameter workflow_file",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]any{
				"owner": "owner",
				"repo":  "repo",
				"ref":   "main",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: workflow_file",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := RunWorkflow(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			require.NoError(t, err)
			require.Equal(t, tc.expectError, result.IsError)

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			if tc.expectedErrMsg != "" {
				assert.Equal(t, tc.expectedErrMsg, textContent.Text)
				return
			}

			// Unmarshal and verify the result
			var response map[string]any
			err = json.Unmarshal([]byte(textContent.Text), &response)
			require.NoError(t, err)
			assert.Equal(t, "Workflow run has been queued", response["message"])
		})
	}
}

func Test_CancelWorkflowRun(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := CancelWorkflowRun(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "cancel_workflow_run", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "run_id")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "run_id"})

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]any
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful workflow run cancellation",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.EndpointPattern{
						Pattern: "/repos/owner/repo/actions/runs/12345/cancel",
						Method:  "POST",
					},
					"", // Empty response body for 202 Accepted
				),
			),
			requestArgs: map[string]any{
				"owner":  "owner",
				"repo":   "repo",
				"run_id": float64(12345),
			},
			expectError: false,
		},
		{
			name:         "missing required parameter run_id",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]any{
				"owner": "owner",
				"repo":  "repo",
			},
			expectError:    false,
			expectedErrMsg: "missing required parameter: run_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := CancelWorkflowRun(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			if tc.expectError {
				require.Error(t, err)
				if tc.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
				return
			}

			require.NoError(t, err)

			// Parse the result and get the text content
			textContent := getTextResult(t, result)

			if tc.expectedErrMsg != "" {
				assert.Contains(t, textContent.Text, tc.expectedErrMsg)
				return
			}

			require.False(t, result.IsError)

			// Unmarshal and verify the result
			var response map[string]any
			err = json.Unmarshal([]byte(textContent.Text), &response)
			require.NoError(t, err)
			assert.Equal(t, "Workflow run has been cancelled", response["message"])
			assert.Equal(t, float64(12345), response["run_id"])
		})
	}
}

func Test_ListWorkflowRunArtifacts(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := ListWorkflowRunArtifacts(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "list_workflow_run_artifacts", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "run_id")
	assert.Contains(t, tool.InputSchema.Properties, "per_page")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "run_id"})

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]any
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful artifacts listing",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposActionsRunsArtifactsByOwnerByRepoByRunId,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						artifacts := &github.ArtifactList{
							TotalCount: github.Int64(2),
							Artifacts: []*github.Artifact{
								{
									ID:                 github.Int64(1),
									NodeID:             github.String("A_1"),
									Name:               github.String("build-artifacts"),
									SizeInBytes:        github.Int64(1024),
									URL:                github.String("https://api.github.com/repos/owner/repo/actions/artifacts/1"),
									ArchiveDownloadURL: github.String("https://api.github.com/repos/owner/repo/actions/artifacts/1/zip"),
									Expired:            github.Bool(false),
									CreatedAt:          &github.Timestamp{},
									UpdatedAt:          &github.Timestamp{},
									ExpiresAt:          &github.Timestamp{},
									WorkflowRun: &github.ArtifactWorkflowRun{
										ID:               github.Int64(12345),
										RepositoryID:     github.Int64(1),
										HeadRepositoryID: github.Int64(1),
										HeadBranch:       github.String("main"),
										HeadSHA:          github.String("abc123"),
									},
								},
								{
									ID:                 github.Int64(2),
									NodeID:             github.String("A_2"),
									Name:               github.String("test-results"),
									SizeInBytes:        github.Int64(512),
									URL:                github.String("https://api.github.com/repos/owner/repo/actions/artifacts/2"),
									ArchiveDownloadURL: github.String("https://api.github.com/repos/owner/repo/actions/artifacts/2/zip"),
									Expired:            github.Bool(false),
									CreatedAt:          &github.Timestamp{},
									UpdatedAt:          &github.Timestamp{},
									ExpiresAt:          &github.Timestamp{},
									WorkflowRun: &github.ArtifactWorkflowRun{
										ID:               github.Int64(12345),
										RepositoryID:     github.Int64(1),
										HeadRepositoryID: github.Int64(1),
										HeadBranch:       github.String("main"),
										HeadSHA:          github.String("abc123"),
									},
								},
							},
						}
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(artifacts)
					}),
				),
			),
			requestArgs: map[string]any{
				"owner":  "owner",
				"repo":   "repo",
				"run_id": float64(12345),
			},
			expectError: false,
		},
		{
			name:         "missing required parameter run_id",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]any{
				"owner": "owner",
				"repo":  "repo",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: run_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := ListWorkflowRunArtifacts(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			require.NoError(t, err)
			require.Equal(t, tc.expectError, result.IsError)

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			if tc.expectedErrMsg != "" {
				assert.Equal(t, tc.expectedErrMsg, textContent.Text)
				return
			}

			// Unmarshal and verify the result
			var response github.ArtifactList
			err = json.Unmarshal([]byte(textContent.Text), &response)
			require.NoError(t, err)
			assert.NotNil(t, response.TotalCount)
			assert.Greater(t, *response.TotalCount, int64(0))
			assert.NotEmpty(t, response.Artifacts)
		})
	}
}

func Test_DownloadWorkflowRunArtifact(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := DownloadWorkflowRunArtifact(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "download_workflow_run_artifact", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "artifact_id")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "artifact_id"})

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]any
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful artifact download URL",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.EndpointPattern{
						Pattern: "/repos/owner/repo/actions/artifacts/123/zip",
						Method:  "GET",
					},
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						// GitHub returns a 302 redirect to the download URL
						w.Header().Set("Location", "https://api.github.com/repos/owner/repo/actions/artifacts/123/download")
						w.WriteHeader(http.StatusFound)
					}),
				),
			),
			requestArgs: map[string]any{
				"owner":       "owner",
				"repo":        "repo",
				"artifact_id": float64(123),
			},
			expectError: false,
		},
		{
			name:         "missing required parameter artifact_id",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]any{
				"owner": "owner",
				"repo":  "repo",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: artifact_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := DownloadWorkflowRunArtifact(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			require.NoError(t, err)
			require.Equal(t, tc.expectError, result.IsError)

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			if tc.expectedErrMsg != "" {
				assert.Equal(t, tc.expectedErrMsg, textContent.Text)
				return
			}

			// Unmarshal and verify the result
			var response map[string]any
			err = json.Unmarshal([]byte(textContent.Text), &response)
			require.NoError(t, err)
			assert.Contains(t, response, "download_url")
			assert.Contains(t, response, "message")
			assert.Equal(t, "Artifact is available for download", response["message"])
			assert.Equal(t, float64(123), response["artifact_id"])
		})
	}
}

func Test_DeleteWorkflowRunLogs(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := DeleteWorkflowRunLogs(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "delete_workflow_run_logs", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "run_id")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "run_id"})

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]any
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful logs deletion",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.DeleteReposActionsRunsLogsByOwnerByRepoByRunId,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNoContent)
					}),
				),
			),
			requestArgs: map[string]any{
				"owner":  "owner",
				"repo":   "repo",
				"run_id": float64(12345),
			},
			expectError: false,
		},
		{
			name:         "missing required parameter run_id",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]any{
				"owner": "owner",
				"repo":  "repo",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: run_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := DeleteWorkflowRunLogs(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			require.NoError(t, err)
			require.Equal(t, tc.expectError, result.IsError)

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			if tc.expectedErrMsg != "" {
				assert.Equal(t, tc.expectedErrMsg, textContent.Text)
				return
			}

			// Unmarshal and verify the result
			var response map[string]any
			err = json.Unmarshal([]byte(textContent.Text), &response)
			require.NoError(t, err)
			assert.Equal(t, "Workflow run logs have been deleted", response["message"])
			assert.Equal(t, float64(12345), response["run_id"])
		})
	}
}

func Test_GetWorkflowRunUsage(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetWorkflowRunUsage(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "get_workflow_run_usage", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "run_id")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "run_id"})

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]any
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful workflow run usage",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposActionsRunsTimingByOwnerByRepoByRunId,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						usage := &github.WorkflowRunUsage{
							Billable: &github.WorkflowRunBillMap{
								"UBUNTU": &github.WorkflowRunBill{
									TotalMS: github.Int64(120000),
									Jobs:    github.Int(2),
									JobRuns: []*github.WorkflowRunJobRun{
										{
											JobID:      github.Int(1),
											DurationMS: github.Int64(60000),
										},
										{
											JobID:      github.Int(2),
											DurationMS: github.Int64(60000),
										},
									},
								},
							},
							RunDurationMS: github.Int64(120000),
						}
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(usage)
					}),
				),
			),
			requestArgs: map[string]any{
				"owner":  "owner",
				"repo":   "repo",
				"run_id": float64(12345),
			},
			expectError: false,
		},
		{
			name:         "missing required parameter run_id",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]any{
				"owner": "owner",
				"repo":  "repo",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: run_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := GetWorkflowRunUsage(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			require.NoError(t, err)
			require.Equal(t, tc.expectError, result.IsError)

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			if tc.expectedErrMsg != "" {
				assert.Equal(t, tc.expectedErrMsg, textContent.Text)
				return
			}

			// Unmarshal and verify the result
			var response github.WorkflowRunUsage
			err = json.Unmarshal([]byte(textContent.Text), &response)
			require.NoError(t, err)
			assert.NotNil(t, response.RunDurationMS)
			assert.NotNil(t, response.Billable)
		})
	}
}
