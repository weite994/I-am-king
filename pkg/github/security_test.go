package github

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetSecuritySettings(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetSecuritySettings(mockClient, translations.NullTranslationHelper)

	assert.Equal(t, "get_security_settings", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})

	// Setup mock security settings
	mockSettings := &github.SecurityAndAnalysis{
		AdvancedSecurity: &github.AdvancedSecurity{
			Status: github.Ptr("enabled"),
		},
		SecretScanning: &github.SecretScanning{
			Status: github.Ptr("enabled"),
		},
		SecretScanningPushProtection: &github.SecretScanningPushProtection{
			Status: github.Ptr("enabled"),
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult *github.SecurityAndAnalysis
		expectedErrMsg string
	}{
		{
			name: "successful security settings retrieval",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposByOwnerByRepo,
					&github.Repository{
						SecurityAndAnalysis: mockSettings,
					},
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
			},
			expectError:    false,
			expectedResult: mockSettings,
		},
		{
			name: "repository not found",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Repository not found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "nonexistent",
			},
			expectError:    true,
			expectedErrMsg: "failed to get repository settings",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := GetSecuritySettings(client, translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedSettings github.SecurityAndAnalysis
			err = json.Unmarshal([]byte(textContent.Text), &returnedSettings)
			require.NoError(t, err)

			assert.Equal(t, *tc.expectedResult.AdvancedSecurity.Status, *returnedSettings.AdvancedSecurity.Status)
			assert.Equal(t, *tc.expectedResult.SecretScanning.Status, *returnedSettings.SecretScanning.Status)
			assert.Equal(t, *tc.expectedResult.SecretScanningPushProtection.Status, *returnedSettings.SecretScanningPushProtection.Status)
		})
	}
}

func Test_GetDependabotSecurityUpdatesStatus(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetDependabotSecurityUpdatesStatus(mockClient, translations.NullTranslationHelper)

	assert.Equal(t, "get_dependabot_security_updates_status", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})

	// Setup mock response
	mockResponse := &github.AutomatedSecurityFixes{
		Enabled: github.Ptr(true),
		Paused:  github.Ptr(false),
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult *github.AutomatedSecurityFixes
		expectedErrMsg string
	}{
		{
			name: "successful status retrieval",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposAutomatedSecurityFixesByOwnerByRepo,
					mockResponse,
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
			},
			expectError:    false,
			expectedResult: mockResponse,
		},
		{
			name: "repository not found",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposAutomatedSecurityFixesByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Repository not found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "nonexistent",
			},
			expectError:    true,
			expectedErrMsg: "failed to get Dependabot security updates status",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := GetDependabotSecurityUpdatesStatus(client, translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedStatus github.AutomatedSecurityFixes
			err = json.Unmarshal([]byte(textContent.Text), &returnedStatus)
			require.NoError(t, err)

			assert.Equal(t, *tc.expectedResult.Enabled, *returnedStatus.Enabled)
			assert.Equal(t, *tc.expectedResult.Paused, *returnedStatus.Paused)
		})
	}
}

func Test_UpdateSecuritySettings(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := UpdateSecuritySettings(mockClient, translations.NullTranslationHelper)

	assert.Equal(t, "update_security_settings", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "settings")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "settings"})

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult *github.SecurityAndAnalysis
		expectedErrMsg string
	}{
		{
			name: "successful update with vulnerability alerts",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposByOwnerByRepo,
					&github.Repository{
						SecurityAndAnalysis: &github.SecurityAndAnalysis{
							AdvancedSecurity: &github.AdvancedSecurity{
								Status: github.Ptr("disabled"),
							},
							SecretScanning: &github.SecretScanning{
								Status: github.Ptr("disabled"),
							},
						},
					},
				),
				mock.WithRequestMatchHandler(
					mock.PatchReposByOwnerByRepo,
					expectRequestBody(t, map[string]interface{}{
						"security_and_analysis": map[string]interface{}{
							"advanced_security": map[string]interface{}{
								"status": "enabled",
							},
							"secret_scanning": map[string]interface{}{
								"status": "disabled",
							},
						},
					}).andThen(
						mockResponse(t, http.StatusOK, &github.Repository{
							SecurityAndAnalysis: &github.SecurityAndAnalysis{
								AdvancedSecurity: &github.AdvancedSecurity{
									Status: github.Ptr("enabled"),
								},
								SecretScanning: &github.SecretScanning{
									Status: github.Ptr("disabled"),
								},
							},
						}),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
				"settings": map[string]interface{}{
					"vulnerability_alerts": true,
				},
			},
			expectError: false,
			expectedResult: &github.SecurityAndAnalysis{
				AdvancedSecurity: &github.AdvancedSecurity{
					Status: github.Ptr("enabled"),
				},
				SecretScanning: &github.SecretScanning{
					Status: github.Ptr("disabled"),
				},
			},
		},
		{
			name: "successful update with multiple settings",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposByOwnerByRepo,
					&github.Repository{
						SecurityAndAnalysis: &github.SecurityAndAnalysis{},
					},
				),
				mock.WithRequestMatchHandler(
					mock.PatchReposByOwnerByRepo,
					expectRequestBody(t, map[string]interface{}{
						"security_and_analysis": map[string]interface{}{
							"advanced_security": map[string]interface{}{
								"status": "enabled",
							},
							"secret_scanning": map[string]interface{}{
								"status": "enabled",
							},
							"secret_scanning_push_protection": map[string]interface{}{
								"status": "enabled",
							},
						},
					}).andThen(
						mockResponse(t, http.StatusOK, &github.Repository{
							SecurityAndAnalysis: &github.SecurityAndAnalysis{
								AdvancedSecurity: &github.AdvancedSecurity{
									Status: github.Ptr("enabled"),
								},
								SecretScanning: &github.SecretScanning{
									Status: github.Ptr("enabled"),
								},
								SecretScanningPushProtection: &github.SecretScanningPushProtection{
									Status: github.Ptr("enabled"),
								},
							},
						}),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
				"settings": map[string]interface{}{
					"advanced_security": map[string]interface{}{
						"status": "enabled",
					},
					"secret_scanning": map[string]interface{}{
						"status": "enabled",
					},
					"secret_scanning_push_protection": map[string]interface{}{
						"status": "enabled",
					},
				},
			},
			expectError: false,
			expectedResult: &github.SecurityAndAnalysis{
				AdvancedSecurity: &github.AdvancedSecurity{
					Status: github.Ptr("enabled"),
				},
				SecretScanning: &github.SecretScanning{
					Status: github.Ptr("enabled"),
				},
				SecretScanningPushProtection: &github.SecretScanningPushProtection{
					Status: github.Ptr("enabled"),
				},
			},
		},
		{
			name: "repository not found",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Repository not found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "nonexistent",
				"settings": map[string]interface{}{
					"vulnerability_alerts": true,
				},
			},
			expectError:    true,
			expectedErrMsg: "failed to get repository settings",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := UpdateSecuritySettings(client, translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedSettings github.SecurityAndAnalysis
			err = json.Unmarshal([]byte(textContent.Text), &returnedSettings)
			require.NoError(t, err)

			if tc.expectedResult.AdvancedSecurity != nil {
				assert.Equal(t, *tc.expectedResult.AdvancedSecurity.Status, *returnedSettings.AdvancedSecurity.Status)
			}
			if tc.expectedResult.SecretScanning != nil {
				assert.Equal(t, *tc.expectedResult.SecretScanning.Status, *returnedSettings.SecretScanning.Status)
			}
			if tc.expectedResult.SecretScanningPushProtection != nil {
				assert.Equal(t, *tc.expectedResult.SecretScanningPushProtection.Status, *returnedSettings.SecretScanningPushProtection.Status)
			}
		})
	}
}

// Test_EnableDependabotSecurityUpdates and Test_DisableDependabotSecurityUpdates are currently disabled.
// See the comment in security.go for details about the GitHub API behavior discrepancy.
// func Test_EnableDependabotSecurityUpdates(t *testing.T) {
// 	// Verify tool definition
// 	mockClient := github.NewClient(nil)
// 	tool, _ := EnableDependabotSecurityUpdates(mockClient, translations.NullTranslationHelper)
//
// 	assert.Equal(t, "enable_dependabot_security_updates", tool.Name)
// 	assert.NotEmpty(t, tool.Description)
// 	assert.Contains(t, tool.InputSchema.Properties, "owner")
// 	assert.Contains(t, tool.InputSchema.Properties, "repo")
// 	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})
//
// 	// Test that the function returns an error indicating the functionality is disabled
// 	_, handler := EnableDependabotSecurityUpdates(mockClient, translations.NullTranslationHelper)
// 	request := createMCPRequest(map[string]interface{}{
// 		"owner": "owner",
// 		"repo":  "repo",
// 	})
//
// 	result, err := handler(context.Background(), request)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "this functionality is currently disabled")
// 	assert.Nil(t, result)
// }
//
// // Test_DisableDependabotSecurityUpdates verifies that the disabled functionality returns an appropriate error
// func Test_DisableDependabotSecurityUpdates(t *testing.T) {
// 	// Verify tool definition
// 	mockClient := github.NewClient(nil)
// 	tool, _ := DisableDependabotSecurityUpdates(mockClient, translations.NullTranslationHelper)
//
// 	assert.Equal(t, "disable_dependabot_security_updates", tool.Name)
// 	assert.NotEmpty(t, tool.Description)
// 	assert.Contains(t, tool.InputSchema.Properties, "owner")
// 	assert.Contains(t, tool.InputSchema.Properties, "repo")
// 	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})
//
// 	// Test that the function returns an error indicating the functionality is disabled
// 	_, handler := DisableDependabotSecurityUpdates(mockClient, translations.NullTranslationHelper)
// 	request := createMCPRequest(map[string]interface{}{
// 		"owner": "owner",
// 		"repo":  "repo",
// 	})
//
// 	result, err := handler(context.Background(), request)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "this functionality is currently disabled")
// 	assert.Nil(t, result)
// }
