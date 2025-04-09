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

func Test_ListRepositories(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := ListRepositories(mockClient, translations.NullTranslationHelper)

	assert.Equal(t, "list_repositories", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "org")
	assert.Contains(t, tool.InputSchema.Properties, "type")
	assert.Contains(t, tool.InputSchema.Properties, "sort")
	assert.Contains(t, tool.InputSchema.Properties, "direction")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"org"})

	// Setup mock repos for success case
	mockRepos := []*github.Repository{
		{
			ID:          github.Ptr(int64(1001)),
			Name:        github.Ptr("repo1"),
			FullName:    github.Ptr("testorg/repo1"),
			Description: github.Ptr("Test repo 1"),
			HTMLURL:     github.Ptr("https://github.com/testorg/repo1"),
			Private:     github.Ptr(false),
			Fork:        github.Ptr(false),
		},
		{
			ID:          github.Ptr(int64(1002)),
			Name:        github.Ptr("repo2"),
			FullName:    github.Ptr("testorg/repo2"),
			Description: github.Ptr("Test repo 2"),
			HTMLURL:     github.Ptr("https://github.com/testorg/repo2"),
			Private:     github.Ptr(true),
			Fork:        github.Ptr(false),
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedRepos  []*github.Repository
		expectedErrMsg string
	}{
		{
			name: "successful repositories listing",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetOrgsReposByOrg,
					expectQueryParams(t, map[string]string{
						"type":      "all",
						"sort":      "created",
						"direction": "desc",
						"per_page":  "30",
						"page":      "1",
					}).andThen(
						mockResponse(t, http.StatusOK, mockRepos),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"org":       "testorg",
				"type":      "all",
				"sort":      "created",
				"direction": "desc",
				"perPage":   float64(30),
				"page":      float64(1),
			},
			expectError:   false,
			expectedRepos: mockRepos,
		},
		{
			name: "successful repos listing with defaults",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetOrgsReposByOrg,
					expectQueryParams(t, map[string]string{
						"per_page": "30",
						"page":     "1",
					}).andThen(
						mockResponse(t, http.StatusOK, mockRepos),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"org": "testorg",
				// Using defaults for other parameters
			},
			expectError:   false,
			expectedRepos: mockRepos,
		},
		{
			name: "custom pagination and filtering",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetOrgsReposByOrg,
					expectQueryParams(t, map[string]string{
						"type":      "public",
						"sort":      "updated",
						"direction": "asc",
						"per_page":  "10",
						"page":      "2",
					}).andThen(
						mockResponse(t, http.StatusOK, mockRepos),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"org":       "testorg",
				"type":      "public",
				"sort":      "updated",
				"direction": "asc",
				"perPage":   float64(10),
				"page":      float64(2),
			},
			expectError:   false,
			expectedRepos: mockRepos,
		},
		{
			name: "API error response",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetOrgsReposByOrg,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"org": "nonexistentorg",
			},
			expectError:    true,
			expectedErrMsg: "failed to list repositories",
		},
		{
			name: "rate limit exceeded",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetOrgsReposByOrg,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusForbidden)
						_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"org": "testorg",
			},
			expectError:    true,
			expectedErrMsg: "failed to list repositories",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := ListRepositories(client, translations.NullTranslationHelper)

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

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedRepos []*github.Repository
			err = json.Unmarshal([]byte(textContent.Text), &returnedRepos)
			require.NoError(t, err)
			assert.Len(t, returnedRepos, len(tc.expectedRepos))
			for i, repo := range returnedRepos {
				assert.Equal(t, *tc.expectedRepos[i].ID, *repo.ID)
				assert.Equal(t, *tc.expectedRepos[i].Name, *repo.Name)
				assert.Equal(t, *tc.expectedRepos[i].FullName, *repo.FullName)
				assert.Equal(t, *tc.expectedRepos[i].Private, *repo.Private)
				assert.Equal(t, *tc.expectedRepos[i].HTMLURL, *repo.HTMLURL)
			}
		})
	}
}
