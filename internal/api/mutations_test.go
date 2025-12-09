package api

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestClientCreateReview(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{
				"addPullRequestReview": {
					"pullRequestReview": {
						"id": "PRR_new123",
						"state": "PENDING"
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		result, err := client.CreateReview(CreateReviewInput{
			PRNodeID:  "PR_abc123",
			CommitOID: "def456",
		})

		if err != nil {
			t.Fatalf("CreateReview() unexpected error: %v", err)
		}
		if result.ID != "PRR_new123" {
			t.Errorf("result.ID = %q, want %q", result.ID, "PRR_new123")
		}
		if result.State != "PENDING" {
			t.Errorf("result.State = %q, want %q", result.State, "PENDING")
		}
	})

	t.Run("empty PR node ID", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.CreateReview(CreateReviewInput{PRNodeID: ""})

		if err == nil {
			t.Error("CreateReview() expected error for empty PR node ID")
		}
		if !strings.Contains(err.Error(), "PR node ID required") {
			t.Errorf("error = %q, want containing 'PR node ID required'", err.Error())
		}
	})

	t.Run("whitespace PR node ID", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.CreateReview(CreateReviewInput{PRNodeID: "   "})

		if err == nil {
			t.Error("CreateReview() expected error for whitespace PR node ID")
		}
	})

	t.Run("GraphQL error", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			return errors.New("mutation failed")
		})

		_, err := client.CreateReview(CreateReviewInput{PRNodeID: "PR_123"})

		if err == nil {
			t.Error("CreateReview() expected error for GraphQL failure")
		}
	})

	t.Run("empty ID in response", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{"addPullRequestReview": {"pullRequestReview": {"id": "", "state": "PENDING"}}}`
			return json.Unmarshal([]byte(resp), response)
		})

		_, err := client.CreateReview(CreateReviewInput{PRNodeID: "PR_123"})

		if err == nil {
			t.Error("CreateReview() expected error for empty ID response")
		}
	})
}

func TestClientAddThread(t *testing.T) {
	t.Run("successful addition", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{
				"addPullRequestReviewThread": {
					"thread": {
						"id": "PRRT_new",
						"path": "main.go",
						"line": 42,
						"isOutdated": false
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		result, err := client.AddThread(AddThreadInput{
			ReviewID: "PRR_123",
			Path:     "main.go",
			Line:     42,
			Body:     "This needs review",
		})

		if err != nil {
			t.Fatalf("AddThread() unexpected error: %v", err)
		}
		if result.ThreadID != "PRRT_new" {
			t.Errorf("ThreadID = %q, want %q", result.ThreadID, "PRRT_new")
		}
		if result.Path != "main.go" {
			t.Errorf("Path = %q, want %q", result.Path, "main.go")
		}
		if result.Line != 42 {
			t.Errorf("Line = %d, want %d", result.Line, 42)
		}
	})

	t.Run("empty review ID", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.AddThread(AddThreadInput{
			ReviewID: "",
			Path:     "file.go",
			Line:     1,
			Body:     "comment",
		})

		if err == nil {
			t.Error("AddThread() expected error for empty review ID")
		}
	})

	t.Run("invalid review ID format", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.AddThread(AddThreadInput{
			ReviewID: "invalid123",
			Path:     "file.go",
			Line:     1,
			Body:     "comment",
		})

		if err == nil {
			t.Error("AddThread() expected error for invalid review ID format")
		}
		if !strings.Contains(err.Error(), "expected GraphQL node ID") {
			t.Errorf("error = %q, want containing 'expected GraphQL node ID'", err.Error())
		}
	})

	t.Run("empty path", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.AddThread(AddThreadInput{
			ReviewID: "PRR_123",
			Path:     "",
			Line:     1,
			Body:     "comment",
		})

		if err == nil {
			t.Error("AddThread() expected error for empty path")
		}
	})

	t.Run("zero line", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.AddThread(AddThreadInput{
			ReviewID: "PRR_123",
			Path:     "file.go",
			Line:     0,
			Body:     "comment",
		})

		if err == nil {
			t.Error("AddThread() expected error for zero line")
		}
	})

	t.Run("negative line", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.AddThread(AddThreadInput{
			ReviewID: "PRR_123",
			Path:     "file.go",
			Line:     -1,
			Body:     "comment",
		})

		if err == nil {
			t.Error("AddThread() expected error for negative line")
		}
	})

	t.Run("empty body", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.AddThread(AddThreadInput{
			ReviewID: "PRR_123",
			Path:     "file.go",
			Line:     1,
			Body:     "",
		})

		if err == nil {
			t.Error("AddThread() expected error for empty body")
		}
	})
}

func TestClientUpdateComment(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{"updatePullRequestReviewComment": {"pullRequestReviewComment": {"id": "PRRC_123"}}}`
			return json.Unmarshal([]byte(resp), response)
		})

		err := client.UpdateComment(UpdateCommentInput{
			CommentID: "PRRC_123",
			Body:      "Updated comment",
		})

		if err != nil {
			t.Fatalf("UpdateComment() unexpected error: %v", err)
		}
	})

	t.Run("empty comment ID", func(t *testing.T) {
		client := newTestClient(nil)
		err := client.UpdateComment(UpdateCommentInput{
			CommentID: "",
			Body:      "comment",
		})

		if err == nil {
			t.Error("UpdateComment() expected error for empty comment ID")
		}
	})

	t.Run("invalid comment ID format", func(t *testing.T) {
		client := newTestClient(nil)
		err := client.UpdateComment(UpdateCommentInput{
			CommentID: "invalid123",
			Body:      "comment",
		})

		if err == nil {
			t.Error("UpdateComment() expected error for invalid comment ID format")
		}
		if !strings.Contains(err.Error(), "expected GraphQL node ID") {
			t.Errorf("error = %q, want containing 'expected GraphQL node ID'", err.Error())
		}
	})

	t.Run("empty body", func(t *testing.T) {
		client := newTestClient(nil)
		err := client.UpdateComment(UpdateCommentInput{
			CommentID: "PRRC_123",
			Body:      "",
		})

		if err == nil {
			t.Error("UpdateComment() expected error for empty body")
		}
	})
}

func TestClientDeleteComment(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{"deletePullRequestReviewComment": {"clientMutationId": null}}`
			return json.Unmarshal([]byte(resp), response)
		})

		err := client.DeleteComment("PRRC_123")

		if err != nil {
			t.Fatalf("DeleteComment() unexpected error: %v", err)
		}
	})

	t.Run("empty comment ID", func(t *testing.T) {
		client := newTestClient(nil)
		err := client.DeleteComment("")

		if err == nil {
			t.Error("DeleteComment() expected error for empty comment ID")
		}
	})

	t.Run("invalid comment ID format", func(t *testing.T) {
		client := newTestClient(nil)
		err := client.DeleteComment("invalid123")

		if err == nil {
			t.Error("DeleteComment() expected error for invalid comment ID format")
		}
	})
}

func TestClientSubmitReview(t *testing.T) {
	t.Run("successful submit", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{"submitPullRequestReview": {"pullRequestReview": {"id": "PRR_123", "state": "APPROVED"}}}`
			return json.Unmarshal([]byte(resp), response)
		})

		err := client.SubmitReview(SubmitReviewInput{
			ReviewID: "PRR_123",
			Event:    "APPROVE",
			Body:     "LGTM!",
		})

		if err != nil {
			t.Fatalf("SubmitReview() unexpected error: %v", err)
		}
	})

	t.Run("empty review ID", func(t *testing.T) {
		client := newTestClient(nil)
		err := client.SubmitReview(SubmitReviewInput{
			ReviewID: "",
			Event:    "APPROVE",
		})

		if err == nil {
			t.Error("SubmitReview() expected error for empty review ID")
		}
	})

	t.Run("invalid review ID format", func(t *testing.T) {
		client := newTestClient(nil)
		err := client.SubmitReview(SubmitReviewInput{
			ReviewID: "invalid123",
			Event:    "APPROVE",
		})

		if err == nil {
			t.Error("SubmitReview() expected error for invalid review ID format")
		}
	})

	t.Run("empty event", func(t *testing.T) {
		client := newTestClient(nil)
		err := client.SubmitReview(SubmitReviewInput{
			ReviewID: "PRR_123",
			Event:    "",
		})

		if err == nil {
			t.Error("SubmitReview() expected error for empty event")
		}
	})
}

func TestClientDeleteReview(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{"deletePullRequestReview": {"clientMutationId": null}}`
			return json.Unmarshal([]byte(resp), response)
		})

		err := client.DeleteReview("PRR_123")

		if err != nil {
			t.Fatalf("DeleteReview() unexpected error: %v", err)
		}
	})

	t.Run("empty review ID", func(t *testing.T) {
		client := newTestClient(nil)
		err := client.DeleteReview("")

		if err == nil {
			t.Error("DeleteReview() expected error for empty review ID")
		}
	})

	t.Run("invalid review ID format", func(t *testing.T) {
		client := newTestClient(nil)
		err := client.DeleteReview("invalid123")

		if err == nil {
			t.Error("DeleteReview() expected error for invalid review ID format")
		}
	})

	t.Run("GraphQL error", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			return errors.New("mutation failed")
		})

		err := client.DeleteReview("PRR_123")

		if err == nil {
			t.Error("DeleteReview() expected error for GraphQL failure")
		}
	})
}
