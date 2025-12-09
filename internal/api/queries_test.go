package api

import (
	"encoding/json"
	"errors"
	"testing"
)

// mockGQLClient implements GraphQLClient interface for testing
type mockGQLClient struct {
	DoFunc func(query string, variables map[string]interface{}, response interface{}) error
}

func (m *mockGQLClient) Do(query string, variables map[string]interface{}, response interface{}) error {
	if m.DoFunc != nil {
		return m.DoFunc(query, variables, response)
	}
	return nil
}

// newTestClient creates a Client with a mock GraphQL client
func newTestClient(doFunc func(query string, variables map[string]interface{}, response interface{}) error) *Client {
	return &Client{gql: &mockGQLClient{DoFunc: doFunc}}
}

func TestNormalizeReviewState(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"PENDING", "pending"},
		{"pending", "pending"},
		{"Pending", "pending"},
		{"APPROVED", "approved"},
		{"approved", "approved"},
		{"CHANGES_REQUESTED", "changes_requested"},
		{"changes_requested", "changes_requested"},
		{"COMMENTED", "commented"},
		{"commented", "commented"},
		{"DISMISSED", "submitted"},
		{"unknown", "submitted"},
		{"", "submitted"},
		{"SUBMITTED", "submitted"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeReviewState(tt.input)
			if got != tt.want {
				t.Errorf("normalizeReviewState(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestClientResolvePR(t *testing.T) {
	t.Run("successful resolution", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			// Simulate GraphQL response
			resp := `{
				"repository": {
					"pullRequest": {
						"id": "PR_123abc",
						"headRefOid": "abc123def456"
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		pr := &PRRef{Owner: "owner", Repo: "repo", Number: 1}
		identity, err := client.ResolvePR(pr)

		if err != nil {
			t.Fatalf("ResolvePR() unexpected error: %v", err)
		}
		if identity.NodeID != "PR_123abc" {
			t.Errorf("NodeID = %q, want %q", identity.NodeID, "PR_123abc")
		}
		if identity.HeadRefOID != "abc123def456" {
			t.Errorf("HeadRefOID = %q, want %q", identity.HeadRefOID, "abc123def456")
		}
	})

	t.Run("PR not found", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{
				"repository": {
					"pullRequest": {
						"id": "",
						"headRefOid": ""
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		pr := &PRRef{Owner: "owner", Repo: "repo", Number: 999}
		_, err := client.ResolvePR(pr)

		if err == nil {
			t.Error("ResolvePR() expected error for empty response")
		}
	})

	t.Run("GraphQL error", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			return errors.New("network error")
		})

		pr := &PRRef{Owner: "owner", Repo: "repo", Number: 1}
		_, err := client.ResolvePR(pr)

		if err == nil {
			t.Error("ResolvePR() expected error for GraphQL failure")
		}
	})
}

func TestClientViewerLogin(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{"viewer": {"login": "testuser"}}`
			return json.Unmarshal([]byte(resp), response)
		})

		login, err := client.ViewerLogin()

		if err != nil {
			t.Fatalf("ViewerLogin() unexpected error: %v", err)
		}
		if login != "testuser" {
			t.Errorf("login = %q, want %q", login, "testuser")
		}
	})

	t.Run("empty login", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{"viewer": {"login": ""}}`
			return json.Unmarshal([]byte(resp), response)
		})

		_, err := client.ViewerLogin()

		if err == nil {
			t.Error("ViewerLogin() expected error for empty login")
		}
	})

	t.Run("whitespace login", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{"viewer": {"login": "   "}}`
			return json.Unmarshal([]byte(resp), response)
		})

		_, err := client.ViewerLogin()

		if err == nil {
			t.Error("ViewerLogin() expected error for whitespace login")
		}
	})
}

func TestClientPendingReviews(t *testing.T) {
	t.Run("returns matching reviews", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{
				"repository": {
					"pullRequest": {
						"reviews": {
							"nodes": [
								{
									"id": "PRR_123",
									"state": "PENDING",
									"url": "https://github.com/owner/repo/pull/1#pullrequestreview-123",
									"updatedAt": "2024-01-15T10:00:00Z",
									"author": {"login": "testuser"},
									"comments": {
										"totalCount": 2,
										"nodes": [
											{"id": "PRRC_1", "path": "file.go", "line": 10, "body": "comment 1", "outdated": false},
											{"id": "PRRC_2", "path": "file.go", "line": 20, "body": "comment 2", "outdated": true}
										]
									}
								}
							]
						}
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		pr := &PRRef{Owner: "owner", Repo: "repo", Number: 1}
		reviews, err := client.PendingReviews(pr, PendingReviewsOptions{Reviewer: "testuser"})

		if err != nil {
			t.Fatalf("PendingReviews() unexpected error: %v", err)
		}
		if len(reviews) != 1 {
			t.Fatalf("PendingReviews() returned %d reviews, want 1", len(reviews))
		}

		review := reviews[0]
		if review.ID != "PRR_123" {
			t.Errorf("review.ID = %q, want %q", review.ID, "PRR_123")
		}
		if len(review.Comments) != 2 {
			t.Errorf("review.Comments length = %d, want 2", len(review.Comments))
		}
	})

	t.Run("filters by reviewer", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{
				"repository": {
					"pullRequest": {
						"reviews": {
							"nodes": [
								{
									"id": "PRR_1",
									"state": "PENDING",
									"url": "",
									"updatedAt": "2024-01-15T10:00:00Z",
									"author": {"login": "user1"},
									"comments": {"totalCount": 0, "nodes": []}
								},
								{
									"id": "PRR_2",
									"state": "PENDING",
									"url": "",
									"updatedAt": "2024-01-15T10:00:00Z",
									"author": {"login": "user2"},
									"comments": {"totalCount": 0, "nodes": []}
								}
							]
						}
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		pr := &PRRef{Owner: "owner", Repo: "repo", Number: 1}
		reviews, err := client.PendingReviews(pr, PendingReviewsOptions{Reviewer: "user1"})

		if err != nil {
			t.Fatalf("PendingReviews() unexpected error: %v", err)
		}
		if len(reviews) != 1 {
			t.Fatalf("PendingReviews() returned %d reviews, want 1 (filtered)", len(reviews))
		}
		if reviews[0].Author != "user1" {
			t.Errorf("review.Author = %q, want %q", reviews[0].Author, "user1")
		}
	})
}

func TestClientAllPRComments(t *testing.T) {
	t.Run("returns review and PR comments", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{
				"repository": {
					"pullRequest": {
						"reviews": {
							"totalCount": 1,
							"nodes": [
								{
									"id": "PRR_1",
									"state": "APPROVED",
									"author": {"login": "reviewer"},
									"comments": {
										"nodes": [
											{"id": "PRRC_1", "path": "main.go", "line": 5, "body": "LGTM", "outdated": false, "author": {"login": "reviewer"}}
										]
									}
								}
							]
						},
						"comments": {
							"totalCount": 1,
							"nodes": [
								{"id": "IC_1", "body": "Discussion comment", "author": {"login": "author"}, "createdAt": "2024-01-15T10:00:00Z"}
							]
						}
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		pr := &PRRef{Owner: "owner", Repo: "repo", Number: 1}
		result, err := client.AllPRComments(pr, AllCommentsOptions{})

		if err != nil {
			t.Fatalf("AllPRComments() unexpected error: %v", err)
		}
		if len(result.ReviewComments) != 1 {
			t.Errorf("ReviewComments length = %d, want 1", len(result.ReviewComments))
		}
		if len(result.PRComments) != 1 {
			t.Errorf("PRComments length = %d, want 1", len(result.PRComments))
		}
		if result.ReviewComments[0].State != "approved" {
			t.Errorf("ReviewComment.State = %q, want %q", result.ReviewComments[0].State, "approved")
		}
	})

	t.Run("truncation flag", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{
				"repository": {
					"pullRequest": {
						"reviews": {"totalCount": 150, "nodes": []},
						"comments": {"totalCount": 50, "nodes": []}
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		pr := &PRRef{Owner: "owner", Repo: "repo", Number: 1}
		result, err := client.AllPRComments(pr, AllCommentsOptions{Limit: 100})

		if err != nil {
			t.Fatalf("AllPRComments() unexpected error: %v", err)
		}
		if !result.Truncated {
			t.Error("expected Truncated = true when totalCount > limit")
		}
	})
}

func TestClientReviewThreads(t *testing.T) {
	t.Run("returns threads", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{
				"repository": {
					"pullRequest": {
						"reviewThreads": {
							"totalCount": 1,
							"nodes": [
								{
									"id": "PRRT_1",
									"isResolved": false,
									"path": "main.go",
									"line": 10,
									"comments": {
										"nodes": [
											{"id": "PRRC_1", "body": "Fix this", "author": {"login": "reviewer"}, "pullRequestReview": {"state": "CHANGES_REQUESTED"}}
										]
									}
								}
							]
						}
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		pr := &PRRef{Owner: "owner", Repo: "repo", Number: 1}
		result, err := client.ReviewThreads(pr, ReviewThreadsOptions{})

		if err != nil {
			t.Fatalf("ReviewThreads() unexpected error: %v", err)
		}
		if len(result.Threads) != 1 {
			t.Fatalf("Threads length = %d, want 1", len(result.Threads))
		}

		thread := result.Threads[0]
		if thread.ID != "PRRT_1" {
			t.Errorf("thread.ID = %q, want %q", thread.ID, "PRRT_1")
		}
		if thread.State != "changes_requested" {
			t.Errorf("thread.State = %q, want %q", thread.State, "changes_requested")
		}
	})

	t.Run("filters unresolved only", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{
				"repository": {
					"pullRequest": {
						"reviewThreads": {
							"totalCount": 2,
							"nodes": [
								{"id": "PRRT_1", "isResolved": false, "path": "a.go", "line": 1, "comments": {"nodes": [{"id": "C1", "body": "open", "author": {"login": "u"}, "pullRequestReview": {"state": "COMMENTED"}}]}},
								{"id": "PRRT_2", "isResolved": true, "path": "b.go", "line": 2, "comments": {"nodes": [{"id": "C2", "body": "closed", "author": {"login": "u"}, "pullRequestReview": {"state": "COMMENTED"}}]}}
							]
						}
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		pr := &PRRef{Owner: "owner", Repo: "repo", Number: 1}
		result, err := client.ReviewThreads(pr, ReviewThreadsOptions{UnresolvedOnly: true})

		if err != nil {
			t.Fatalf("ReviewThreads() unexpected error: %v", err)
		}
		if len(result.Threads) != 1 {
			t.Fatalf("Threads length = %d, want 1 (filtered)", len(result.Threads))
		}
		if result.Threads[0].ID != "PRRT_1" {
			t.Errorf("expected unresolved thread PRRT_1")
		}
	})

	t.Run("filters by states", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{
				"repository": {
					"pullRequest": {
						"reviewThreads": {
							"totalCount": 2,
							"nodes": [
								{"id": "PRRT_1", "isResolved": false, "path": "a.go", "line": 1, "comments": {"nodes": [{"id": "C1", "body": "pending", "author": {"login": "u"}, "pullRequestReview": {"state": "PENDING"}}]}},
								{"id": "PRRT_2", "isResolved": false, "path": "b.go", "line": 2, "comments": {"nodes": [{"id": "C2", "body": "approved", "author": {"login": "u"}, "pullRequestReview": {"state": "APPROVED"}}]}}
							]
						}
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		pr := &PRRef{Owner: "owner", Repo: "repo", Number: 1}
		result, err := client.ReviewThreads(pr, ReviewThreadsOptions{States: []string{"pending"}})

		if err != nil {
			t.Fatalf("ReviewThreads() unexpected error: %v", err)
		}
		if len(result.Threads) != 1 {
			t.Fatalf("Threads length = %d, want 1 (filtered by state)", len(result.Threads))
		}
		if result.Threads[0].State != "pending" {
			t.Errorf("expected pending state thread")
		}
	})
}

func TestLatestPendingReview(t *testing.T) {
	t.Run("returns most recent", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{
				"repository": {
					"pullRequest": {
						"reviews": {
							"nodes": [
								{"id": "PRR_old", "state": "PENDING", "url": "", "updatedAt": "2024-01-01T10:00:00Z", "author": {"login": "user"}, "comments": {"totalCount": 0, "nodes": []}},
								{"id": "PRR_new", "state": "PENDING", "url": "", "updatedAt": "2024-01-15T10:00:00Z", "author": {"login": "user"}, "comments": {"totalCount": 0, "nodes": []}}
							]
						}
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		pr := &PRRef{Owner: "owner", Repo: "repo", Number: 1}
		review, err := client.LatestPendingReview(pr, PendingReviewsOptions{Reviewer: "user"})

		if err != nil {
			t.Fatalf("LatestPendingReview() unexpected error: %v", err)
		}
		if review.ID != "PRR_new" {
			t.Errorf("review.ID = %q, want %q (most recent)", review.ID, "PRR_new")
		}
	})

	t.Run("no pending reviews", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{
				"repository": {
					"pullRequest": {
						"reviews": {"nodes": []}
					}
				}
			}`
			return json.Unmarshal([]byte(resp), response)
		})

		pr := &PRRef{Owner: "owner", Repo: "repo", Number: 1}
		_, err := client.LatestPendingReview(pr, PendingReviewsOptions{Reviewer: "user"})

		if err == nil {
			t.Error("LatestPendingReview() expected error when no reviews")
		}
	})
}
