package api

import (
	"fmt"
	"strings"
)

type CreateReviewInput struct {
	PRNodeID  string
	CommitOID string
}

type CreateReviewResult struct {
	ID    string
	State string
}

func (c *Client) CreateReview(input CreateReviewInput) (*CreateReviewResult, error) {
	prNodeID := strings.TrimSpace(input.PRNodeID)
	if prNodeID == "" {
		return nil, fmt.Errorf("PR node ID required")
	}

	commitOID := strings.TrimSpace(input.CommitOID)

	const mutation = `mutation CreateReview($input: AddPullRequestReviewInput!) {
  addPullRequestReview(input: $input) {
    pullRequestReview {
      id
      state
    }
  }
}`

	mutationInput := map[string]interface{}{
		"pullRequestId": prNodeID,
	}
	if commitOID != "" {
		mutationInput["commitOID"] = commitOID
	}

	variables := map[string]interface{}{
		"input": mutationInput,
	}

	var response struct {
		AddPullRequestReview struct {
			PullRequestReview struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"pullRequestReview"`
		} `json:"addPullRequestReview"`
	}

	if err := c.gql.Do(mutation, variables, &response); err != nil {
		return nil, fmt.Errorf("create review: %w", err)
	}

	id := strings.TrimSpace(response.AddPullRequestReview.PullRequestReview.ID)
	if id == "" {
		return nil, fmt.Errorf("create review returned empty ID")
	}

	return &CreateReviewResult{
		ID:    id,
		State: response.AddPullRequestReview.PullRequestReview.State,
	}, nil
}

type AddThreadInput struct {
	ReviewID  string
	Path      string
	Line      int
	Side      string
	Body      string
	StartLine *int
	StartSide *string
}

type AddThreadResult struct {
	ThreadID string
	Path     string
	Line     int
	Outdated bool
}

func (c *Client) AddThread(input AddThreadInput) (*AddThreadResult, error) {
	reviewID := strings.TrimSpace(input.ReviewID)
	if reviewID == "" {
		return nil, fmt.Errorf("review ID required")
	}
	if !strings.HasPrefix(reviewID, "PRR_") {
		return nil, fmt.Errorf("invalid review ID %q: expected GraphQL node ID", reviewID)
	}

	path := strings.TrimSpace(input.Path)
	if path == "" {
		return nil, fmt.Errorf("path required")
	}
	if input.Line <= 0 {
		return nil, fmt.Errorf("line must be positive")
	}

	body := strings.TrimSpace(input.Body)
	if body == "" {
		return nil, fmt.Errorf("body required")
	}

	side := strings.ToUpper(strings.TrimSpace(input.Side))
	if side == "" {
		side = "RIGHT"
	}

	const mutation = `mutation AddThread($input: AddPullRequestReviewThreadInput!) {
  addPullRequestReviewThread(input: $input) {
    thread {
      id
      path
      line
      isOutdated
    }
  }
}`

	mutationInput := map[string]interface{}{
		"pullRequestReviewId": reviewID,
		"path":                path,
		"line":                input.Line,
		"side":                side,
		"body":                body,
	}
	if input.StartLine != nil {
		mutationInput["startLine"] = *input.StartLine
	}
	if input.StartSide != nil {
		mutationInput["startSide"] = strings.ToUpper(*input.StartSide)
	}

	variables := map[string]interface{}{
		"input": mutationInput,
	}

	var response struct {
		AddPullRequestReviewThread struct {
			Thread struct {
				ID         string `json:"id"`
				Path       string `json:"path"`
				Line       *int   `json:"line"`
				IsOutdated bool   `json:"isOutdated"`
			} `json:"thread"`
		} `json:"addPullRequestReviewThread"`
	}

	if err := c.gql.Do(mutation, variables, &response); err != nil {
		return nil, fmt.Errorf("add thread: %w", err)
	}

	threadID := strings.TrimSpace(response.AddPullRequestReviewThread.Thread.ID)
	if threadID == "" {
		return nil, fmt.Errorf("add thread returned empty ID")
	}

	line := 0
	if response.AddPullRequestReviewThread.Thread.Line != nil {
		line = *response.AddPullRequestReviewThread.Thread.Line
	}

	return &AddThreadResult{
		ThreadID: threadID,
		Path:     response.AddPullRequestReviewThread.Thread.Path,
		Line:     line,
		Outdated: response.AddPullRequestReviewThread.Thread.IsOutdated,
	}, nil
}

type UpdateCommentInput struct {
	CommentID string
	Body      string
}

func (c *Client) UpdateComment(input UpdateCommentInput) error {
	commentID := strings.TrimSpace(input.CommentID)
	if commentID == "" {
		return fmt.Errorf("comment ID required")
	}
	if !strings.HasPrefix(commentID, "PRRC_") {
		return fmt.Errorf("invalid comment ID %q: expected GraphQL node ID", commentID)
	}

	body := strings.TrimSpace(input.Body)
	if body == "" {
		return fmt.Errorf("body required")
	}

	const mutation = `mutation UpdateComment($input: UpdatePullRequestReviewCommentInput!) {
  updatePullRequestReviewComment(input: $input) {
    pullRequestReviewComment {
      id
    }
  }
}`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"pullRequestReviewCommentId": commentID,
			"body":                        body,
		},
	}

	var response struct {
		UpdatePullRequestReviewComment struct {
			PullRequestReviewComment struct {
				ID string `json:"id"`
			} `json:"pullRequestReviewComment"`
		} `json:"updatePullRequestReviewComment"`
	}

	if err := c.gql.Do(mutation, variables, &response); err != nil {
		return fmt.Errorf("update comment: %w", err)
	}

	return nil
}

func (c *Client) DeleteComment(commentID string) error {
	commentID = strings.TrimSpace(commentID)
	if commentID == "" {
		return fmt.Errorf("comment ID required")
	}
	if !strings.HasPrefix(commentID, "PRRC_") {
		return fmt.Errorf("invalid comment ID %q: expected GraphQL node ID", commentID)
	}

	const mutation = `mutation DeleteComment($input: DeletePullRequestReviewCommentInput!) {
  deletePullRequestReviewComment(input: $input) {
    clientMutationId
  }
}`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"id": commentID,
		},
	}

	var response struct {
		DeletePullRequestReviewComment struct {
			ClientMutationID *string `json:"clientMutationId"`
		} `json:"deletePullRequestReviewComment"`
	}

	if err := c.gql.Do(mutation, variables, &response); err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	return nil
}

type SubmitReviewInput struct {
	ReviewID string
	Event    string
	Body     string
}

func (c *Client) SubmitReview(input SubmitReviewInput) error {
	reviewID := strings.TrimSpace(input.ReviewID)
	if reviewID == "" {
		return fmt.Errorf("review ID required")
	}
	if !strings.HasPrefix(reviewID, "PRR_") {
		return fmt.Errorf("invalid review ID %q: expected GraphQL node ID", reviewID)
	}

	event := strings.ToUpper(strings.TrimSpace(input.Event))
	if event == "" {
		return fmt.Errorf("event required")
	}

	const mutation = `mutation SubmitReview($input: SubmitPullRequestReviewInput!) {
  submitPullRequestReview(input: $input) {
    pullRequestReview {
      id
      state
    }
  }
}`

	mutationInput := map[string]interface{}{
		"pullRequestReviewId": reviewID,
		"event":               event,
	}
	if body := strings.TrimSpace(input.Body); body != "" {
		mutationInput["body"] = body
	}

	variables := map[string]interface{}{
		"input": mutationInput,
	}

	var response struct {
		SubmitPullRequestReview struct {
			PullRequestReview struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"pullRequestReview"`
		} `json:"submitPullRequestReview"`
	}

	if err := c.gql.Do(mutation, variables, &response); err != nil {
		return fmt.Errorf("submit review: %w", err)
	}

	return nil
}

func (c *Client) DeleteReview(reviewID string) error {
	reviewID = strings.TrimSpace(reviewID)
	if reviewID == "" {
		return fmt.Errorf("review ID required")
	}
	if !strings.HasPrefix(reviewID, "PRR_") {
		return fmt.Errorf("invalid review ID %q: expected GraphQL node ID", reviewID)
	}

	const mutation = `mutation DeleteReview($input: DeletePullRequestReviewInput!) {
  deletePullRequestReview(input: $input) {
    clientMutationId
  }
}`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"pullRequestReviewId": reviewID,
		},
	}

	var response struct {
		DeletePullRequestReview struct {
			ClientMutationID *string `json:"clientMutationId"`
		} `json:"deletePullRequestReview"`
	}

	if err := c.gql.Do(mutation, variables, &response); err != nil {
		return fmt.Errorf("delete review: %w", err)
	}

	return nil
}
