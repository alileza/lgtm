package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofri/go-github-ratelimit/v2/github_ratelimit"
	"github.com/google/go-github/v75/github"
	"golang.org/x/oauth2"
)

// GitHubClient handles GitHub API operations
type GitHubClient struct {
	client *github.Client
	config *Configuration
}

// ApprovalRequest represents a request to approve a GitHub pull request
type ApprovalRequest struct {
	Owner         string
	Repository    string
	PRNumber      int
	Message       string
	SourceChannel string
	SourceUser    string
	SourceMessage *SlackMessage
	Timestamp     time.Time
}

// ApprovalResult represents the result of a GitHub PR approval operation
type ApprovalResult struct {
	Request        *ApprovalRequest
	Success        bool
	ReviewID       int64
	Error          string
	ProcessedAt    time.Time
	RetryAttempts  int
}

// NewGitHubClient creates a new GitHub client with rate limiting
func NewGitHubClient(config *Configuration) (*GitHubClient, error) {
	// Create OAuth2 token source
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.GitHubToken})
	
	// Create OAuth2 HTTP client
	oauthClient := oauth2.NewClient(ctx, ts)
	
	// Create rate-limited HTTP client
	rateLimitedClient := github_ratelimit.NewClient(oauthClient.Transport)
	
	// Create GitHub client
	client := github.NewClient(rateLimitedClient)
	
	return &GitHubClient{
		client: client,
		config: config,
	}, nil
}

// ValidatePermissions checks if the GitHub token has required permissions
func (gc *GitHubClient) ValidatePermissions(ctx context.Context) error {
	// Test basic authentication by getting the authenticated user
	user, _, err := gc.client.Users.Get(ctx, "")
	if err != nil {
		return &AuthenticationError{Service: "GitHub", Message: fmt.Sprintf("authentication failed: %v", err)}
	}
	
	logInfo("Authenticated as GitHub user: %s", user.GetLogin())
	
	// Test if we can access the repository (if default repo is configured)
	if gc.config.DefaultOwner != "" && gc.config.DefaultRepo != "" {
		_, _, err := gc.client.Repositories.Get(ctx, gc.config.DefaultOwner, gc.config.DefaultRepo)
		if err != nil {
			return &AuthenticationError{
				Service: "GitHub", 
				Message: fmt.Sprintf("insufficient permissions for repository %s/%s: %v", 
					gc.config.DefaultOwner, gc.config.DefaultRepo, err),
			}
		}
		
		logInfo("Repository access confirmed: %s/%s", gc.config.DefaultOwner, gc.config.DefaultRepo)
	}
	
	return nil
}

// GetAuthenticatedUser returns information about the authenticated user
func (gc *GitHubClient) GetAuthenticatedUser(ctx context.Context) (*github.User, error) {
	user, _, err := gc.client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user: %v", err)
	}
	return user, nil
}

// ValidatePRReference checks if a PR exists and is in a valid state for approval
func (gc *GitHubClient) ValidatePRReference(ctx context.Context, owner, repo string, prNumber int) error {
	logDebug("Validating PR: %s/%s#%d", owner, repo, prNumber)
	
	// Get the pull request
	pr, response, err := gc.client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		if response != nil {
			switch response.StatusCode {
			case 404:
				return fmt.Errorf("PR #%d not found in %s/%s", prNumber, owner, repo)
			case 403:
				return fmt.Errorf("insufficient permissions to access PR #%d in %s/%s", prNumber, owner, repo)
			}
		}
		return fmt.Errorf("failed to get PR #%d: %v", prNumber, err)
	}
	
	// Check if PR is in a valid state for approval
	if pr.GetState() != "open" {
		return fmt.Errorf("PR #%d is %s and cannot be approved", prNumber, pr.GetState())
	}
	
	if pr.GetMerged() {
		return fmt.Errorf("PR #%d is already merged", prNumber)
	}
	
	logDebug("PR validation successful: %s/%s#%d state=%s mergeable=%v", owner, repo, prNumber, pr.GetState(), pr.GetMergeable())
	
	return nil
}

// ApprovePR approves a GitHub pull request
func (gc *GitHubClient) ApprovePR(ctx context.Context, req *ApprovalRequest) (*ApprovalResult, error) {
	result := &ApprovalResult{
		Request:     req,
		ProcessedAt: time.Now(),
	}
	
	logDebug("Approving PR: %s/%s#%d", req.Owner, req.Repository, req.PRNumber)
	
	// Create review request with approval
	reviewRequest := &github.PullRequestReviewRequest{
		Event: github.String("APPROVE"),
	}
	
	// Submit the review
	review, response, err := gc.client.PullRequests.CreateReview(
		ctx,
		req.Owner,
		req.Repository,
		req.PRNumber,
		reviewRequest,
	)
	
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to approve PR #%d: %v", req.PRNumber, err)
		
		// Check if it's a specific error we can handle
		if response != nil {
			switch response.StatusCode {
			case 404:
				result.Error = fmt.Sprintf("PR #%d not found in %s/%s", req.PRNumber, req.Owner, req.Repository)
			case 403:
				result.Error = fmt.Sprintf("insufficient permissions to approve PR #%d", req.PRNumber)
			case 422:
				result.Error = fmt.Sprintf("PR #%d cannot be approved (already merged or closed)", req.PRNumber)
			}
		}
		
		return result, nil // Return result with error, don't fail the function
	}
	
	// Success
	result.Success = true
	result.ReviewID = review.GetID()
	
	logDebug("PR approved successfully: %s/%s#%d review_id=%d", req.Owner, req.Repository, req.PRNumber, result.ReviewID)
	
	return result, nil
}

// ApprovePRWithRetry approves a GitHub PR with retry logic
func (gc *GitHubClient) ApprovePRWithRetry(ctx context.Context, req *ApprovalRequest) (*ApprovalResult, error) {
	const maxRetries = 3
	const baseDelay = time.Second
	
	var lastResult *ApprovalResult
	var lastErr error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 2^attempt seconds
			delay := time.Duration(1<<uint(attempt)) * baseDelay
			logDebug("Retrying PR approval: attempt=%d/%d delay=%v pr_number=%d", attempt+1, maxRetries, delay, req.PRNumber)
			
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		
		result, err := gc.ApprovePR(ctx, req)
		if err != nil {
			lastErr = err
			lastResult = result
			logDebug("PR approval attempt failed: attempt=%d error=%v", attempt+1, err)
			continue
		}
		
		// If the result indicates success or a permanent failure, don't retry
		if result.Success || isPermanentError(result.Error) {
			if result.RetryAttempts == 0 {
				result.RetryAttempts = attempt
			}
			return result, nil
		}
		
		// Store the last result for potential retry
		lastResult = result
		lastErr = fmt.Errorf("approval failed: %s", result.Error)
	}
	
	// All retries exhausted
	if lastResult != nil {
		lastResult.RetryAttempts = maxRetries
		logWarn("PR approval retries exhausted: %s/%s#%d final_error=%s", req.Owner, req.Repository, req.PRNumber, lastResult.Error)
		return lastResult, nil
	}
	
	return nil, fmt.Errorf("approval failed after %d attempts: %v", maxRetries, lastErr)
}

// isPermanentError determines if an error should not be retried
func isPermanentError(errorMsg string) bool {
	permanentErrors := []string{
		"not found",
		"insufficient permissions",
		"already merged",
		"already closed",
		"invalid_auth",
		"Bad credentials",
	}
	
	for _, permanent := range permanentErrors {
		if strings.Contains(errorMsg, permanent) {
			return true
		}
	}
	
	return false
}