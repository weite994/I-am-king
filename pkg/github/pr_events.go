package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// PREventContext holds common state for PR event handlers
type PREventContext struct {
	MCPServer     *server.MCPServer
	Client        *github.Client
	Ctx           context.Context
	Request       mcp.CallToolRequest
	Owner         string
	Repo          string
	PullNumber    int
	TimeoutSecs   int
	HasTimeout    bool
	StartTime     time.Time
	ProgressToken any
	PollInterval  time.Duration
}

// handleResponse is a helper function to handle GitHub API responses and properly close the body
func handleResponse(resp *github.Response, errorPrefix string) error {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		return fmt.Errorf("%s: %s", errorPrefix, string(body))
	}
	return nil
}

// sendProgressNotification sends a progress notification to the client
func sendProgressNotification(ctx context.Context, eventCtx *PREventContext) {
	if eventCtx.ProgressToken == nil {
		return
	}

	// Calculate elapsed time
	elapsed := time.Since(eventCtx.StartTime)

	// Calculate progress value
	var progress float64
	var total *float64
	if eventCtx.HasTimeout {
		// If timeout is set, use percentage of elapsed time
		// Get the deadline from the context
		deadline, _ := ctx.Deadline()
		timeoutDuration := deadline.Sub(eventCtx.StartTime)
		progress = elapsed.Seconds() / timeoutDuration.Seconds()
		totalValue := 1.0
		total = &totalValue
	} else {
		// If no timeout, just increment progress endlessly
		progress = elapsed.Seconds()
		// No total value when incrementing endlessly
		total = nil
	}

	// Create and send a progress notification with the client's token
	n := mcp.NewProgressNotification(eventCtx.ProgressToken, progress, total)

	// Send the progress notification to the client
	params := map[string]any{"progressToken": n.Params.ProgressToken, "progress": n.Params.Progress, "total": n.Params.Total}
	if err := eventCtx.MCPServer.SendNotificationToClient(ctx, "notifications/progress", params); err != nil {
		// Log the error but continue - notification failures shouldn't stop the process
		fmt.Printf("Failed to send progress notification: %v\n", err)
	}
}

// parsePREventParams parses common parameters for PR event handlers and sets up the context
func parsePREventParams(ctx context.Context, mcpServer *server.MCPServer, client *github.Client, request mcp.CallToolRequest) (*PREventContext, *mcp.CallToolResult, context.CancelFunc, error) {
	eventCtx := &PREventContext{
		MCPServer:    mcpServer,
		Client:       client,
		Ctx:          ctx,
		Request:      request,
		PollInterval: 5 * time.Second,
		StartTime:    time.Now(),
	}

	// Get required parameters
	owner, err := requiredParam[string](request, "owner")
	if err != nil {
		return nil, mcp.NewToolResultError(err.Error()), nil, nil
	}
	eventCtx.Owner = owner

	repo, err := requiredParam[string](request, "repo")
	if err != nil {
		return nil, mcp.NewToolResultError(err.Error()), nil, nil
	}
	eventCtx.Repo = repo

	pullNumber, err := requiredInt(request, "pullNumber")
	if err != nil {
		return nil, mcp.NewToolResultError(err.Error()), nil, nil
	}
	eventCtx.PullNumber = pullNumber

	// Get timeout parameter
	timeoutSecs, err := optionalIntParam(request, "timeout_seconds")
	if err != nil {
		return nil, mcp.NewToolResultError(err.Error()), nil, nil
	}
	eventCtx.TimeoutSecs = timeoutSecs

	// If timeout is provided, create a child context with timeout
	eventCtx.HasTimeout = timeoutSecs > 0
	var cancel context.CancelFunc
	if eventCtx.HasTimeout {
		eventCtx.Ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
		// We can't defer cancel() here because this function returns before the context should be canceled
		// The caller must handle cancellation
	} else {
		cancel = func() {} // No-op cancel function if no timeout
	}

	// Extract the client's progress token (if any)
	if request.Params.Meta != nil {
		eventCtx.ProgressToken = request.Params.Meta.ProgressToken
	}

	return eventCtx, nil, cancel, nil
}

// pollForPREvent runs a polling loop for PR events with proper context handling
func pollForPREvent(eventCtx *PREventContext, checkFn func() (*mcp.CallToolResult, error)) (*mcp.CallToolResult, error) {
	// Use a defer to ensure we send a final progress update if needed
	defer func() {
		if eventCtx.ProgressToken != nil {
			sendProgressNotification(eventCtx.Ctx, eventCtx)
		}
	}()

	// Enter polling loop
	for {
		// Check if context is done (canceled or deadline exceeded)
		select {
		case <-eventCtx.Ctx.Done():
			if eventCtx.HasTimeout && eventCtx.Ctx.Err() == context.DeadlineExceeded {
				// Customize the timeout message based on the tool name
				var operation string
				if strings.Contains(eventCtx.Request.Method, "wait_for_pr_checks") {
					operation = "PR checks to complete"
				} else if strings.Contains(eventCtx.Request.Method, "wait_for_pr_review") {
					operation = "PR review"
				} else {
					operation = "operation"
				}
				return mcp.NewToolResultError(fmt.Sprintf("Timeout waiting for %s after %d seconds", operation, eventCtx.TimeoutSecs)), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("Operation canceled: %v", eventCtx.Ctx.Err())), nil
		default:
			// Continue with current logic
		}

		// Call the check function
		result, err := checkFn()
		// nil will be returned for result AND nil when we have not yet completed
		// our check
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}

		// Send progress notification
		sendProgressNotification(eventCtx.Ctx, eventCtx)

		// Sleep before next poll
		time.Sleep(eventCtx.PollInterval)
	}
}

// waitForPRChecks creates a tool to wait for all status checks to complete on a pull request.
func waitForPRChecks(mcpServer *server.MCPServer, client *github.Client, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("wait_for_pr_checks",
			mcp.WithDescription(t("TOOL_WAIT_FOR_PR_CHECKS_DESCRIPTION", "Wait for all status checks to complete on a pull request")),
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
			mcp.WithNumber("timeout_seconds",
				mcp.Description("How long to wait before giving up. When not provided, no timeout will be applied."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Parse common parameters and set up context
			eventCtx, result, cancel, err := parsePREventParams(ctx, mcpServer, client, request)
			if result != nil || err != nil {
				return result, err
			}
			defer cancel()

			// Run the polling loop with a check function for PR checks
			return pollForPREvent(eventCtx, func() (*mcp.CallToolResult, error) {
				// First get the PR to find the head SHA
				pr, resp, err := client.PullRequests.Get(eventCtx.Ctx, eventCtx.Owner, eventCtx.Repo, eventCtx.PullNumber)
				if err != nil {
					return nil, fmt.Errorf("failed to get pull request: %w", err)
				}

				// Handle the response
				if err := handleResponse(resp, "failed to get pull request"); err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				// Get combined status for the head SHA
				status, resp, err := client.Repositories.GetCombinedStatus(eventCtx.Ctx, eventCtx.Owner, eventCtx.Repo, *pr.Head.SHA, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to get combined status: %w", err)
				}

				// Handle the response
				if err := handleResponse(resp, "failed to get combined status"); err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				// Check if all checks are complete
				if status.State != nil && (*status.State == "success" || *status.State == "failure" || *status.State == "error") {
					// All checks are complete, return the status
					r, err := json.Marshal(status)
					if err != nil {
						return nil, fmt.Errorf("failed to marshal response: %w", err)
					}
					return mcp.NewToolResultText(string(r)), nil
				}

				// Return nil to continue polling
				return nil, nil
			})
		}
}

// waitForPRReview creates a tool to wait for a new review to be added to a pull request.
func waitForPRReview(mcpServer *server.MCPServer, client *github.Client, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("wait_for_pr_review",
			mcp.WithDescription(t("TOOL_WAIT_FOR_PR_REVIEW_DESCRIPTION", "Wait for a new review to be added to a pull request")),
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
			mcp.WithNumber("timeout_seconds",
				mcp.Description("How long to wait before giving up. When not provided, no timeout will be applied."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Parse common parameters and set up context
			eventCtx, result, cancel, err := parsePREventParams(ctx, mcpServer, client, request)
			if result != nil || err != nil {
				return result, err
			}
			defer cancel()

			// Get optional last_review_id parameter
			lastReviewID, err := optionalIntParam(request, "last_review_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Run the polling loop with a check function for PR reviews
			return pollForPREvent(eventCtx, func() (*mcp.CallToolResult, error) {
				// Get the current reviews
				reviews, resp, err := client.PullRequests.ListReviews(eventCtx.Ctx, eventCtx.Owner, eventCtx.Repo, eventCtx.PullNumber, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to get pull request reviews: %w", err)
				}

				// Handle the response
				if err := handleResponse(resp, "failed to get pull request reviews"); err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				// Check if there are any new reviews
				var latestReview *github.PullRequestReview
				for _, review := range reviews {
					if review.ID == nil {
						continue
					}

					reviewID := int(*review.ID)
					if reviewID > lastReviewID {
						if latestReview == nil || reviewID > int(*latestReview.ID) {
							latestReview = review
						}
					}
				}

				// If we found a new review, return it
				if latestReview != nil {
					r, err := json.Marshal(latestReview)
					if err != nil {
						return nil, fmt.Errorf("failed to marshal response: %w", err)
					}
					return mcp.NewToolResultText(string(r)), nil
				}

				// Return nil to continue polling
				return nil, nil
			})
		}
}
