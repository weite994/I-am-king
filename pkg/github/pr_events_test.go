package github

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/server"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_WaitForPRChecks(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	mockServer := server.NewMCPServer("test", "1.0.0")
	tool, _ := waitForPRChecks(mockServer, mockClient, translations.NullTranslationHelper)

	assert.Equal(t, "wait_for_pr_checks", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})
	// timeout_seconds parameter has been removed

	// Setup mock PR for successful PR fetch
	mockPR := &github.PullRequest{
		Number:  github.Ptr(42),
		Title:   github.Ptr("Test PR"),
		HTMLURL: github.Ptr("https://github.com/owner/repo/pull/42"),
		Head: &github.PullRequestBranch{
			SHA: github.Ptr("abcd1234"),
			Ref: github.Ptr("feature-branch"),
		},
	}

	// Setup mock check runs for completed case
	mockCompletedCheckRuns := &github.ListCheckRunsResults{
		Total: github.Ptr(2),
		CheckRuns: []*github.CheckRun{
			{
				ID:         github.Ptr(int64(1)),
				Name:       github.Ptr("travis-ci"),
				Status:     github.Ptr("completed"),
				Conclusion: github.Ptr("success"),
				HTMLURL:    github.Ptr("https://travis-ci.org/owner/repo/builds/123"),
				Output: &github.CheckRunOutput{
					Title:   github.Ptr("Build succeeded"),
					Summary: github.Ptr("All tests passed"),
				},
			},
			{
				ID:         github.Ptr(int64(2)),
				Name:       github.Ptr("codecov"),
				Status:     github.Ptr("completed"),
				Conclusion: github.Ptr("success"),
				HTMLURL:    github.Ptr("https://codecov.io/gh/owner/repo/pull/42"),
				Output: &github.CheckRunOutput{
					Title:   github.Ptr("Coverage increased"),
					Summary: github.Ptr("Coverage is now at 85%"),
				},
			},
		},
	}

	// Setup mock check runs for in-progress case
	mockInProgressCheckRuns := &github.ListCheckRunsResults{
		Total: github.Ptr(2),
		CheckRuns: []*github.CheckRun{
			{
				ID:         github.Ptr(int64(1)),
				Name:       github.Ptr("travis-ci"),
				Status:     github.Ptr("completed"),
				Conclusion: github.Ptr("success"),
				HTMLURL:    github.Ptr("https://travis-ci.org/owner/repo/builds/123"),
				Output: &github.CheckRunOutput{
					Title:   github.Ptr("Build succeeded"),
					Summary: github.Ptr("All tests passed"),
				},
			},
			{
				ID:      github.Ptr(int64(2)),
				Name:    github.Ptr("codecov"),
				Status:  github.Ptr("in_progress"),
				HTMLURL: github.Ptr("https://codecov.io/gh/owner/repo/pull/42"),
				Output: &github.CheckRunOutput{
					Title:   github.Ptr("Coverage in progress"),
					Summary: github.Ptr("Calculating coverage"),
				},
			},
		},
	}

	// Setup mock empty check runs
	mockEmptyCheckRuns := &github.ListCheckRunsResults{
		Total:     github.Ptr(0),
		CheckRuns: []*github.CheckRun{},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]any
		expectError    bool
		expectProgress bool
		expectedStatus *github.ListCheckRunsResults
		expectedErrMsg string
	}{
		{
			name: "checks completed successfully",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposPullsByOwnerByRepoByPullNumber,
					mockPR,
				),
				mock.WithRequestMatch(
					mock.GetReposCommitsCheckRunsByOwnerByRepoByRef,
					mockCompletedCheckRuns,
				),
			),
			requestArgs: map[string]any{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError:    false,
			expectProgress: false,
			expectedStatus: mockCompletedCheckRuns,
		},
		{
			name: "checks still pending",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposPullsByOwnerByRepoByPullNumber,
					mockPR,
				),
				mock.WithRequestMatch(
					mock.GetReposCommitsCheckRunsByOwnerByRepoByRef,
					mockInProgressCheckRuns,
				),
			),
			requestArgs: map[string]any{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError:    true,
			expectedErrMsg: "Timeout waiting for",
		},
		{
			name: "no check runs configured",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposPullsByOwnerByRepoByPullNumber,
					mockPR,
				),
				mock.WithRequestMatch(
					mock.GetReposCommitsCheckRunsByOwnerByRepoByRef,
					mockEmptyCheckRuns,
				),
			),
			requestArgs: map[string]any{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError:    false,
			expectProgress: false,
			expectedStatus: mockEmptyCheckRuns,
		},
		{
			name: "PR fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposPullsByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]any{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(999),
			},
			expectError:    true,
			expectedErrMsg: "failed to get pull request",
		},
		{
			name: "status fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposPullsByOwnerByRepoByPullNumber,
					mockPR,
				),
				mock.WithRequestMatchHandler(
					mock.GetReposCommitsStatusByOwnerByRepoByRef,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]any{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError:    true,
			expectedErrMsg: "failed to get check runs",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			mockServer := server.NewMCPServer("test", "1.0.0")
			_, handler := waitForPRChecks(mockServer, client, translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Create a context with timeout to prevent tests from running too long
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// Call handler
			result, err := handler(ctx, request)

			// Verify results
			if tc.expectError {
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
					return
				}
				// Error might be in the result instead
				textContent := getTextResult(t, result)
				assert.Contains(t, textContent.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// For completed responses, unmarshal and verify the check runs
			var returnedCheckRuns github.ListCheckRunsResults
			err = json.Unmarshal([]byte(textContent.Text), &returnedCheckRuns)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedStatus.Total, *returnedCheckRuns.Total)
			assert.Len(t, returnedCheckRuns.CheckRuns, len(tc.expectedStatus.CheckRuns))
			// Verify the first check run
			if len(returnedCheckRuns.CheckRuns) > 0 && len(tc.expectedStatus.CheckRuns) > 0 {
				assert.Equal(t, *tc.expectedStatus.CheckRuns[0].Name, *returnedCheckRuns.CheckRuns[0].Name)
				assert.Equal(t, *tc.expectedStatus.CheckRuns[0].Status, *returnedCheckRuns.CheckRuns[0].Status)
			}
		})
	}
}

func Test_WaitForPRReview(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	// Create a mock githubv4.Client
	mockGQLClient := &githubv4.Client{}
	mockServer := server.NewMCPServer("test", "1.0.0")
	tool, _ := waitForPRReview(mockServer, mockClient, mockGQLClient, translations.NullTranslationHelper)

	assert.Equal(t, "wait_for_pr_review", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.Contains(t, tool.InputSchema.Properties, "last_review_id")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})
	// timeout_seconds parameter has been removed

	// Setup mock PR reviews for success case
	mockReviews := []*github.PullRequestReview{
		{
			ID:      github.Ptr(int64(201)),
			State:   github.Ptr("APPROVED"),
			Body:    github.Ptr("LGTM"),
			HTMLURL: github.Ptr("https://github.com/owner/repo/pull/42#pullrequestreview-201"),
			User: &github.User{
				Login: github.Ptr("approver"),
			},
			CommitID:    github.Ptr("abcdef123456"),
			SubmittedAt: &github.Timestamp{Time: time.Now().Add(-24 * time.Hour)},
		},
		{
			ID:      github.Ptr(int64(202)),
			State:   github.Ptr("CHANGES_REQUESTED"),
			Body:    github.Ptr("Please address the following issues"),
			HTMLURL: github.Ptr("https://github.com/owner/repo/pull/42#pullrequestreview-202"),
			User: &github.User{
				Login: github.Ptr("reviewer"),
			},
			CommitID:    github.Ptr("abcdef123456"),
			SubmittedAt: &github.Timestamp{Time: time.Now().Add(-12 * time.Hour)},
		},
		{
			ID:      github.Ptr(int64(203)),
			State:   github.Ptr("APPROVED"),
			Body:    github.Ptr("Now it looks good!"),
			HTMLURL: github.Ptr("https://github.com/owner/repo/pull/42#pullrequestreview-203"),
			User: &github.User{
				Login: github.Ptr("reviewer"),
			},
			CommitID:    github.Ptr("abcdef789012"),
			SubmittedAt: &github.Timestamp{Time: time.Now().Add(-1 * time.Hour)},
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]any
		expectError    bool
		expectProgress bool
		expectedReview *github.PullRequestReview
		expectedErrMsg string
	}{
		{
			name: "new review found",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposPullsReviewsByOwnerByRepoByPullNumber,
					mockReviews,
				),
			),
			requestArgs: map[string]any{
				"owner":          "owner",
				"repo":           "repo",
				"pullNumber":     float64(42),
				"last_review_id": float64(202),
			},
			expectError:    false,
			expectProgress: false,
			expectedReview: mockReviews[2], // The newest review (ID 203)
		},
		{
			name: "no new reviews",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposPullsReviewsByOwnerByRepoByPullNumber,
					mockReviews,
				),
			),
			requestArgs: map[string]any{
				"owner":          "owner",
				"repo":           "repo",
				"pullNumber":     float64(42),
				"last_review_id": float64(203), // Already have the latest review
			},
			expectError:    true,
			expectedErrMsg: "Timeout waiting for",
		},

		{
			name: "reviews fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposPullsReviewsByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]any{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(999),
			},
			expectError:    true,
			expectedErrMsg: "failed to get pull request reviews",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			// Create a mock githubv4.Client for each test case
			mockGQLClient := &githubv4.Client{}
			mockServer := server.NewMCPServer("test", "1.0.0")
			_, handler := waitForPRReview(mockServer, client, mockGQLClient, translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Create a context with timeout to prevent tests from running too long
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// Call handler
			result, err := handler(ctx, request)

			// Verify results
			if tc.expectError {
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
					return
				}
				// Error might be in the result instead
				textContent := getTextResult(t, result)
				assert.Contains(t, textContent.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// For completed responses, unmarshal and verify the review
			var returnedReview github.PullRequestReview
			err = json.Unmarshal([]byte(textContent.Text), &returnedReview)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedReview.ID, *returnedReview.ID)
			assert.Equal(t, *tc.expectedReview.State, *returnedReview.State)
			assert.Equal(t, *tc.expectedReview.Body, *returnedReview.Body)
			assert.Equal(t, *tc.expectedReview.User.Login, *returnedReview.User.Login)
		})
	}
}
