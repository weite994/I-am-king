package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shurcooL/githubv4"
)

// PullRequestEventContext holds common state for pull request event handlers
type PullRequestEventContext struct {
	MCPServer     *server.MCPServer
	Client        *github.Client
	Ctx           context.Context
	Request       mcp.CallToolRequest
	Owner         string
	Repo          string
	PullNumber    int
	StartTime     time.Time
	ProgressToken any
	PollInterval  time.Duration
}

// handleResponse is a helper function to handle GitHub API responses and properly close the body
func handleResponse(resp *github.Response, errorPrefix string) error {
	fmt.Fprintf(os.Stderr, "[DEBUG] handleResponse: Processing response with status code %d for %s\n", resp.StatusCode, errorPrefix)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "[DEBUG] handleResponse: Non-OK status code %d detected\n", resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] handleResponse: Failed to read response body: %v\n", err)
			return fmt.Errorf("failed to read response body: %w", err)
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] handleResponse: Error response body: %s\n", string(body))
		return fmt.Errorf("%s: %s", errorPrefix, string(body))
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] handleResponse: Successfully processed response\n")
	return nil
}

// sendProgressNotification sends a progress notification to the client
func sendProgressNotification(ctx context.Context, eventCtx *PullRequestEventContext) {
	fmt.Fprintf(os.Stderr, "[DEBUG] sendProgressNotification: Starting progress notification process\n")
	if eventCtx.ProgressToken == nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] sendProgressNotification: No progress token, skipping notification\n")
		return
	}

	// Calculate elapsed time
	elapsed := time.Since(eventCtx.StartTime)
	fmt.Fprintf(os.Stderr, "[DEBUG] sendProgressNotification: Elapsed time: %v\n", elapsed)

	// Calculate progress value
	var progress float64
	var total *float64
	// Just increment progress endlessly
	progress = elapsed.Seconds()
	// No total value when incrementing endlessly
	total = nil
	fmt.Fprintf(os.Stderr, "[DEBUG] sendProgressNotification: Progress without timeout: %f (no total)\n", progress)

	// Create and send a progress notification with the client's token
	n := mcp.NewProgressNotification(eventCtx.ProgressToken, progress, total)
	fmt.Fprintf(os.Stderr, "[DEBUG] sendProgressNotification: Created notification with token: %v\n", eventCtx.ProgressToken)

	// Send the progress notification to the client
	params := map[string]any{"progressToken": n.Params.ProgressToken, "progress": n.Params.Progress, "total": n.Params.Total}
	fmt.Fprintf(os.Stderr, "[DEBUG] sendProgressNotification: Sending notification with params: %+v\n", params)
	if err := eventCtx.MCPServer.SendNotificationToClient(ctx, "notifications/progress", params); err != nil {
		// Log the error but continue - notification failures shouldn't stop the process
		fmt.Printf("Failed to send progress notification: %v\n", err)
		fmt.Fprintf(os.Stderr, "[DEBUG] sendProgressNotification: ERROR sending notification: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG] sendProgressNotification: Successfully sent notification\n")
	}
}

// parsePullRequestEventParams parses common parameters for pull request event handlers and sets up the context
func parsePullRequestEventParams(ctx context.Context, mcpServer *server.MCPServer, client *github.Client, request mcp.CallToolRequest) (*PullRequestEventContext, *mcp.CallToolResult, context.CancelFunc, error) {
	fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: Starting parameter parsing for method: %s\n", request.Method)
	eventCtx := &PullRequestEventContext{
		MCPServer:    mcpServer,
		Client:       client,
		Ctx:          ctx,
		Request:      request,
		PollInterval: 5 * time.Second,
		StartTime:    time.Now(),
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: Created event context with poll interval: %v\n", eventCtx.PollInterval)

	// Get required parameters
	fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: Extracting required 'owner' parameter\n")
	owner, err := requiredParam[string](request, "owner")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: ERROR extracting 'owner' parameter: %v\n", err)
		return nil, mcp.NewToolResultError(err.Error()), nil, nil
	}
	eventCtx.Owner = owner
	fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: Owner: %s\n", owner)

	fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: Extracting required 'repo' parameter\n")
	repo, err := requiredParam[string](request, "repo")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: ERROR extracting 'repo' parameter: %v\n", err)
		return nil, mcp.NewToolResultError(err.Error()), nil, nil
	}
	eventCtx.Repo = repo
	fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: Repo: %s\n", repo)

	fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: Extracting required 'pullNumber' parameter\n")
	pullNumber, err := requiredInt(request, "pullNumber")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: ERROR extracting 'pullNumber' parameter: %v\n", err)
		return nil, mcp.NewToolResultError(err.Error()), nil, nil
	}
	eventCtx.PullNumber = pullNumber
	fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: Pull Number: %d\n", pullNumber)

	// Create a no-op cancel function
	var cancel context.CancelFunc = func() {} // No-op cancel function
	fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: Using context without timeout\n")

	// Extract the client's progress token (if any)
	fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: Checking for progress token\n")
	if request.Params.Meta != nil {
		eventCtx.ProgressToken = request.Params.Meta.ProgressToken
		fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: Progress token found: %v\n", eventCtx.ProgressToken)
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: No progress token found\n")
	}

	fmt.Fprintf(os.Stderr, "[DEBUG] parsePullRequestEventParams: Parameter parsing complete\n")
	return eventCtx, nil, cancel, nil
}

// pollForPullRequestEvent runs a polling loop for pull request events with proper context handling
func pollForPullRequestEvent(eventCtx *PullRequestEventContext, checkFn func() (*mcp.CallToolResult, error)) (*mcp.CallToolResult, error) {
	fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Starting polling loop for method: %s\n", eventCtx.Request.Method)
	// Use a defer to ensure we send a final progress update if needed
	defer func() {
		fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Exiting polling loop, sending final progress update\n")
		if eventCtx.ProgressToken != nil {
			sendProgressNotification(eventCtx.Ctx, eventCtx)
		} else {
			fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: No progress token, skipping final update\n")
		}
	}()

	// Enter polling loop
	pollCount := 0
	for {
		pollCount++
		fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Poll iteration #%d\n", pollCount)

		// Check if context is done (canceled or deadline exceeded)
		select {
		case <-eventCtx.Ctx.Done():
			fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Context done with error: %v\n", eventCtx.Ctx.Err())
			if eventCtx.Ctx.Err() == context.DeadlineExceeded {
				// Customize the timeout message based on the tool name
				var operation string
				switch {
				case strings.Contains(eventCtx.Request.Method, "wait_for_pullrequest_checks"):
					operation = "pull request checks to complete"
				case strings.Contains(eventCtx.Request.Method, "wait_for_pullrequest_review"):
					operation = "pull request review"
				default:
					operation = "operation"
				}
				fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Timeout exceeded waiting for %s\n", operation)
				return mcp.NewToolResultError(fmt.Sprintf("Timeout waiting for %s", operation)), nil
			}
			fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Operation canceled: %v\n", eventCtx.Ctx.Err())
			return mcp.NewToolResultError(fmt.Sprintf("Operation canceled: %v", eventCtx.Ctx.Err())), nil
		default:
			// Continue with current logic
			fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Context still active, continuing\n")
		}

		// Call the check function
		fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Calling check function\n")
		result, err := checkFn()
		// nil will be returned for result AND nil when we have not yet completed
		// our check
		if err != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Check function returned error: %v\n", err)
			return nil, err
		}
		if result != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Check function returned result, exiting poll loop\n")
			return result, nil
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Check function returned no result, continuing polling\n")

		// Send progress notification
		fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Sending progress notification\n")
		sendProgressNotification(eventCtx.Ctx, eventCtx)

		// Sleep before next poll
		fmt.Fprintf(os.Stderr, "[DEBUG] pollForPullRequestEvent: Sleeping for %v before next poll\n", eventCtx.PollInterval)
		time.Sleep(eventCtx.PollInterval)
	}
}

// waitForPullRequestReview creates a tool to wait for a new review to be added to a pull request.
func waitForPullRequestReview(mcpServer *server.MCPServer, gh *github.Client, gql *githubv4.Client, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview: Registering wait_for_pullrequest_review tool\n")
	return mcp.NewTool("wait_for_pullrequest_review",
			mcp.WithDescription(t("TOOL_WAIT_FOR_PULLREQUEST_REVIEW_DESCRIPTION", "Wait for a pull request to be approved, or for additional feedback to be added")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithNumber("pullNumber",
				mcp.Required(),
				mcp.Description("Pull request number"),
			),
			mcp.WithNumber("last_review_id",
				mcp.Description("ID of most recent review (wait for newer reviews)"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview: Tool handler called with method: %s\n", request.Method)
			// Parse common parameters and set up context
			eventCtx, result, cancel, err := parsePullRequestEventParams(ctx, mcpServer, gh, request)
			if result != nil || err != nil {
				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview: Parameter parsing failed: %v\n", err)
				return result, err
			}
			defer cancel()
			fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview: Parameters parsed successfully\n")

			// Get optional last_review_id parameter
			fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview: Extracting optional 'last_review_id' parameter\n")
			lastReviewID, err := optionalIntParam(request, "last_review_id")
			if err != nil {
				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview: ERROR extracting 'last_review_id' parameter: %v\n", err)
				return mcp.NewToolResultError(err.Error()), nil
			}
			fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview: Last review ID: %d\n", lastReviewID)
			fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview: Starting polling for new reviews\n")

			// Run the polling loop with a check function for pull request reviews
			return pollForPullRequestEvent(eventCtx, func() (*mcp.CallToolResult, error) {
				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: Checking for new reviews\n")
				// Get the current reviews
				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: Listing reviews for %s/%s #%d\n", eventCtx.Owner, eventCtx.Repo, eventCtx.PullNumber)
				reviews, resp, err := gh.PullRequests.ListReviews(eventCtx.Ctx, eventCtx.Owner, eventCtx.Repo, eventCtx.PullNumber, nil)
				if err != nil {
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: ERROR listing reviews: %v\n", err)
					return nil, fmt.Errorf("failed to get pull request reviews: %w", err)
				}

				// Handle the response
				if err := handleResponse(resp, "failed to get pull request reviews"); err != nil {
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: ERROR handling reviews response: %v\n", err)
					return mcp.NewToolResultError(err.Error()), nil
				}

				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: Found %d reviews, checking for new ones after ID %d\n", len(reviews), lastReviewID)
				// Check if there are any new reviews
				var latestReview *github.PullRequestReview
				for _, review := range reviews {
					if review.ID == nil {
						fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: Skipping review with nil ID\n")
						continue
					}

					reviewID := int(*review.ID)
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: Checking review ID: %d\n", reviewID)
					if reviewID > lastReviewID {
						fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: Found newer review with ID: %d\n", reviewID)
						if latestReview == nil || reviewID > int(*latestReview.ID) {
							fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: This is the latest review so far\n")
							latestReview = review
						}
					}
				}

				// If we found a new review, return it
				if latestReview != nil {
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: Found new review with ID: %d\n", *latestReview.ID)
					r, err := json.Marshal(latestReview)
					if err != nil {
						fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: ERROR marshaling review: %v\n", err)
						return nil, fmt.Errorf("failed to marshal response: %w", err)
					}
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: Returning new review result\n")
					return mcp.NewToolResultText(string(r)), nil
				}

				// Return nil to continue polling
				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestReview.checkFn: No new reviews found, continuing polling\n")
				return nil, nil
			})
		}
}

// waitForPullRequestChecks creates a tool to wait for all status checks to complete on a pull request.
func waitForPullRequestChecks(mcpServer *server.MCPServer, client *github.Client, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks: Registering wait_for_pullrequest_checks tool\n")
	return mcp.NewTool("wait_for_pullrequest_checks",
			mcp.WithDescription(t("TOOL_WAIT_FOR_PULLREQUEST_CHECKS_DESCRIPTION", "Wait for all status checks to complete on a pull request")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithNumber("pullNumber",
				mcp.Required(),
				mcp.Description("Pull request number"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks: Tool handler called with method: %s\n", request.Method)
			// Parse common parameters and set up context
			eventCtx, result, cancel, err := parsePullRequestEventParams(ctx, mcpServer, client, request)
			if result != nil || err != nil {
				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks: Parameter parsing failed: %v\n", err)
				return result, err
			}
			defer cancel()
			fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks: Parameters parsed successfully, starting polling\n")

			// Run the polling loop with a check function for pull request checks
			return pollForPullRequestEvent(eventCtx, func() (*mcp.CallToolResult, error) {
				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: Checking pull request status\n")
				// First get the pull request to find the head SHA
				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: Getting pull request details for %s/%s #%d\n", eventCtx.Owner, eventCtx.Repo, eventCtx.PullNumber)
				pr, resp, err := client.PullRequests.Get(eventCtx.Ctx, eventCtx.Owner, eventCtx.Repo, eventCtx.PullNumber)
				if err != nil {
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: ERROR getting pull request: %v\n", err)
					return nil, fmt.Errorf("failed to get pull request: %w", err)
				}

				// Handle the response
				if err := handleResponse(resp, "failed to get pull request"); err != nil {
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: ERROR handling pull request response: %v\n", err)
					return mcp.NewToolResultError(err.Error()), nil
				}

				if pr.Head == nil || pr.Head.SHA == nil {
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: ERROR: Pull request head or SHA is nil\n")
					return mcp.NewToolResultError("Pull request head SHA is missing"), nil
				}

				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: Got pull request head SHA: %s\n", *pr.Head.SHA)

				// Get check runs for the head SHA
				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: Getting check runs for SHA: %s\n", *pr.Head.SHA)
				checkRuns, resp, err := client.Checks.ListCheckRunsForRef(eventCtx.Ctx, eventCtx.Owner, eventCtx.Repo, *pr.Head.SHA, nil)

				if err != nil {
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: ERROR getting check runs: %v\n", err)
					return nil, fmt.Errorf("failed to get check runs: %w", err)
				}

				// Handle the response
				if err := handleResponse(resp, "failed to get check runs"); err != nil {
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: ERROR handling check runs response: %v\n", err)
					return mcp.NewToolResultError(err.Error()), nil
				}

				// Check if there are any check runs
				if checkRuns.GetTotal() == 0 {
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: No check runs found, considering checks complete\n")
					// If there are no check runs, we should consider the checks complete
					// Otherwise, we'd poll indefinitely for repositories without checks
					r, err := json.Marshal(checkRuns)
					if err != nil {
						fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: ERROR marshaling empty check runs: %v\n", err)
						return nil, fmt.Errorf("failed to marshal response: %w", err)
					}
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: Returning empty check runs result\n")
					return mcp.NewToolResultText(string(r)), nil
				}

				// Check if all checks are complete
				allComplete := true
				for _, checkRun := range checkRuns.CheckRuns {
					if checkRun.GetStatus() != "completed" {
						allComplete = false
						break
					}
				}

				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: All checks complete: %v, Total checks: %d\n", allComplete, checkRuns.GetTotal())
				if allComplete {
					// All checks are complete, return the check runs
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: All checks complete\n")
					r, err := json.Marshal(checkRuns)
					if err != nil {
						fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: ERROR marshaling response: %v\n", err)
						return nil, fmt.Errorf("failed to marshal response: %w", err)
					}
					fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: Returning successful result\n")
					return mcp.NewToolResultText(string(r)), nil
				}

				// Return nil to continue polling
				fmt.Fprintf(os.Stderr, "[DEBUG] waitForPullRequestChecks.checkFn: Checks still in progress, continuing polling\n")
				return nil, nil
			})
		}
}
