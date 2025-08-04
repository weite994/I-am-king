package github

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/github/github-mcp-server/internal/toolsnaps"
	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v73/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ListStarredRepositories(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := ListStarredRepositories(stubGetClientFn(mockClient), translations.NullTranslationHelper)
	require.NoError(t, toolsnaps.Test(tool.Name, tool))

	assert.Equal(t, "list_starred_repositories", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "username")
	assert.Contains(t, tool.InputSchema.Properties, "sort")
	assert.Contains(t, tool.InputSchema.Properties, "direction")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"username"})

	// Setup mock starred repositories for success case
	mockStarredRepos := []*github.StarredRepository{
		{
			Repository: &github.Repository{
				ID:       github.Ptr(int64(1)),
				Name:     github.Ptr("awesome-repo"),
				FullName: github.Ptr("owner/awesome-repo"),
				Owner: &github.User{
					Login: github.Ptr("owner"),
				},
				Description:     github.Ptr("An awesome repository"),
				HTMLURL:         github.Ptr("https://github.com/owner/awesome-repo"),
				StargazersCount: github.Ptr(100),
				Language:        github.Ptr("Go"),
				Fork:            github.Ptr(false),
				Private:         github.Ptr(false),
			},
			StarredAt: &github.Timestamp{Time: time.Now().Add(-24 * time.Hour)},
		},
		{
			Repository: &github.Repository{
				ID:       github.Ptr(int64(2)),
				Name:     github.Ptr("cool-project"),
				FullName: github.Ptr("another/cool-project"),
				Owner: &github.User{
					Login: github.Ptr("another"),
				},
				Description:     github.Ptr("A cool project"),
				HTMLURL:         github.Ptr("https://github.com/another/cool-project"),
				StargazersCount: github.Ptr(250),
				Language:        github.Ptr("JavaScript"),
				Fork:            github.Ptr(true),
				Private:         github.Ptr(false),
			},
			StarredAt: &github.Timestamp{Time: time.Now().Add(-48 * time.Hour)},
		},
	}

	tests := []struct {
		name                 string
		mockedClient         *http.Client
		requestArgs          map[string]interface{}
		expectError          bool
		expectedStarredRepos []*github.StarredRepository
		expectedErrMsg       string
	}{
		{
			name: "successful starred repositories retrieval with minimal parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetUsersStarredByUsername,
					mockStarredRepos,
				),
			),
			requestArgs: map[string]interface{}{
				"username": "testuser",
			},
			expectError:          false,
			expectedStarredRepos: mockStarredRepos,
		},
		{
			name: "successful starred repositories retrieval with all parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetUsersStarredByUsername,
					expectQueryParams(t, map[string]string{
						"sort":      "created",
						"direction": "desc",
						"page":      "2",
						"per_page":  "50",
					}).andThen(
						mockResponse(t, http.StatusOK, mockStarredRepos),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"username":  "testuser",
				"sort":      "created",
				"direction": "desc",
				"page":      float64(2),
				"perPage":   float64(50),
			},
			expectError:          false,
			expectedStarredRepos: mockStarredRepos,
		},
		{
			name: "successful starred repositories retrieval with sort updated",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetUsersStarredByUsername,
					expectQueryParams(t, map[string]string{
						"sort":      "updated",
						"direction": "asc",
						"page":      "1",
						"per_page":  "30",
					}).andThen(
						mockResponse(t, http.StatusOK, mockStarredRepos),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"username":  "testuser",
				"sort":      "updated",
				"direction": "asc",
			},
			expectError:          false,
			expectedStarredRepos: mockStarredRepos,
		},
		{
			name: "successful starred repositories retrieval with default perPage",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetUsersStarredByUsername,
					expectQueryParams(t, map[string]string{
						"page":     "1",
						"per_page": "30", // Default perPage should be 30
					}).andThen(
						mockResponse(t, http.StatusOK, mockStarredRepos),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"username": "testuser",
				"page":     float64(1),
			},
			expectError:          false,
			expectedStarredRepos: mockStarredRepos,
		},
		{
			name: "empty starred repositories list",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetUsersStarredByUsername,
					[]*github.StarredRepository{},
				),
			),
			requestArgs: map[string]interface{}{
				"username": "userwithnorepos",
			},
			expectError:          false,
			expectedStarredRepos: []*github.StarredRepository{},
		},
		{
			name: "user not found",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetUsersStarredByUsername,
					mockResponse(t, http.StatusNotFound, `{"message": "Not Found"}`),
				),
			),
			requestArgs: map[string]interface{}{
				"username": "nonexistentuser",
			},
			expectError:    true,
			expectedErrMsg: "failed to list starred repositories for user",
		},
		{
			name: "rate limit exceeded",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetUsersStarredByUsername,
					mockResponse(t, http.StatusForbidden, `{"message": "API rate limit exceeded"}`),
				),
			),
			requestArgs: map[string]interface{}{
				"username": "testuser",
			},
			expectError:    true,
			expectedErrMsg: "failed to list starred repositories for user",
		},
		{
			name:           "missing required parameter username",
			mockedClient:   mock.NewMockedHTTPClient(),
			requestArgs:    map[string]interface{}{},
			expectError:    false,
			expectedErrMsg: "missing required parameter: username",
		},
		{
			name:         "invalid sort parameter",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"username": "testuser",
				"sort":     "invalid",
			},
			expectError:    false,
			expectedErrMsg: "invalid value for sort parameter",
		},
		{
			name:         "invalid direction parameter",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"username":  "testuser",
				"direction": "invalid",
			},
			expectError:    false,
			expectedErrMsg: "invalid value for direction parameter",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := ListStarredRepositories(stubGetClientFn(client), translations.NullTranslationHelper)

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

			if tc.expectedErrMsg != "" {
				require.NotNil(t, result)
				textContent := getTextResult(t, result)
				assert.Contains(t, textContent.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedStarredRepos []*github.StarredRepository
			err = json.Unmarshal([]byte(textContent.Text), &returnedStarredRepos)
			require.NoError(t, err)

			assert.Len(t, returnedStarredRepos, len(tc.expectedStarredRepos))
			for i, starredRepo := range returnedStarredRepos {
				if i < len(tc.expectedStarredRepos) {
					assert.Equal(t, *tc.expectedStarredRepos[i].Repository.ID, *starredRepo.Repository.ID)
					assert.Equal(t, *tc.expectedStarredRepos[i].Repository.Name, *starredRepo.Repository.Name)
					assert.Equal(t, *tc.expectedStarredRepos[i].Repository.FullName, *starredRepo.Repository.FullName)
					assert.Equal(t, *tc.expectedStarredRepos[i].Repository.Owner.Login, *starredRepo.Repository.Owner.Login)
					assert.Equal(t, *tc.expectedStarredRepos[i].Repository.HTMLURL, *starredRepo.Repository.HTMLURL)
					assert.NotNil(t, starredRepo.StarredAt)

					if tc.expectedStarredRepos[i].Repository.Description != nil {
						assert.Equal(t, *tc.expectedStarredRepos[i].Repository.Description, *starredRepo.Repository.Description)
					}
					if tc.expectedStarredRepos[i].Repository.Language != nil {
						assert.Equal(t, *tc.expectedStarredRepos[i].Repository.Language, *starredRepo.Repository.Language)
					}
				}
			}
		})
	}
}

func Test_ListStarredRepositories_ParameterValidation(t *testing.T) {
	// Test parameter validation without making HTTP requests
	mockClient := github.NewClient(nil)
	_, handler := ListStarredRepositories(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	tests := []struct {
		name        string
		requestArgs map[string]interface{}
		expectedErr string
	}{
		{
			name: "valid sort parameter - created",
			requestArgs: map[string]interface{}{
				"username": "testuser",
				"sort":     "created",
			},
			expectedErr: "",
		},
		{
			name: "valid sort parameter - updated",
			requestArgs: map[string]interface{}{
				"username": "testuser",
				"sort":     "updated",
			},
			expectedErr: "",
		},
		{
			name: "valid direction parameter - asc",
			requestArgs: map[string]interface{}{
				"username":  "testuser",
				"direction": "asc",
			},
			expectedErr: "",
		},
		{
			name: "valid direction parameter - desc",
			requestArgs: map[string]interface{}{
				"username":  "testuser",
				"direction": "desc",
			},
			expectedErr: "",
		},
		{
			name: "invalid sort parameter",
			requestArgs: map[string]interface{}{
				"username": "testuser",
				"sort":     "popularity",
			},
			expectedErr: "invalid value for sort parameter",
		},
		{
			name: "invalid direction parameter",
			requestArgs: map[string]interface{}{
				"username":  "testuser",
				"direction": "random",
			},
			expectedErr: "invalid value for direction parameter",
		},
		{
			name: "negative page parameter",
			requestArgs: map[string]interface{}{
				"username": "testuser",
				"page":     float64(-1),
			},
			expectedErr: "page must be greater than 0",
		},
		{
			name: "zero page parameter",
			requestArgs: map[string]interface{}{
				"username": "testuser",
				"page":     float64(0),
			},
			expectedErr: "page must be greater than 0",
		},
		{
			name: "negative perPage parameter",
			requestArgs: map[string]interface{}{
				"username": "testuser",
				"perPage":  float64(-1),
			},
			expectedErr: "perPage must be between 1 and 100",
		},
		{
			name: "zero perPage parameter",
			requestArgs: map[string]interface{}{
				"username": "testuser",
				"perPage":  float64(0),
			},
			expectedErr: "perPage must be between 1 and 100",
		},
		{
			name: "perPage parameter too large",
			requestArgs: map[string]interface{}{
				"username": "testuser",
				"perPage":  float64(101),
			},
			expectedErr: "perPage must be between 1 and 100",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			if tc.expectedErr != "" {
				// Should return a tool error, not a Go error
				require.NoError(t, err)
				require.NotNil(t, result)
				textContent := getTextResult(t, result)
				assert.Contains(t, textContent.Text, tc.expectedErr)
			} else if err != nil {
				// Should not return an error for valid parameters
				// (though it might fail due to network issues in this test setup)
				// If there's an error, it should be a network/client error, not a validation error
				assert.NotContains(t, err.Error(), "invalid value")
				assert.NotContains(t, err.Error(), "must be")
			}
		})
	}
}

func Test_ListStarredRepositories_DefaultPerPage(t *testing.T) {
	// Test that default perPage is set to 30 when not provided
	mockStarredRepos := []*github.StarredRepository{
		{
			Repository: &github.Repository{
				ID:       github.Ptr(int64(1)),
				Name:     github.Ptr("test-repo"),
				FullName: github.Ptr("owner/test-repo"),
				Owner: &github.User{
					Login: github.Ptr("owner"),
				},
				HTMLURL: github.Ptr("https://github.com/owner/test-repo"),
			},
			StarredAt: &github.Timestamp{Time: time.Now()},
		},
	}

	mockedClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetUsersStarredByUsername,
			expectQueryParams(t, map[string]string{
				"page":     "1",
				"per_page": "30", // Should default to 30
			}).andThen(
				mockResponse(t, http.StatusOK, mockStarredRepos),
			),
		),
	)

	client := github.NewClient(mockedClient)
	_, handler := ListStarredRepositories(stubGetClientFn(client), translations.NullTranslationHelper)

	// Create call request without perPage parameter
	request := createMCPRequest(map[string]interface{}{
		"username": "testuser",
	})

	// Call handler
	result, err := handler(context.Background(), request)

	require.NoError(t, err)
	textContent := getTextResult(t, result)

	var returnedStarredRepos []*github.StarredRepository
	err = json.Unmarshal([]byte(textContent.Text), &returnedStarredRepos)
	require.NoError(t, err)
	assert.Len(t, returnedStarredRepos, 1)
}
