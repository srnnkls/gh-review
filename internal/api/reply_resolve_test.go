package api

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestClientReplyThread(t *testing.T) {
	t.Run("successful reply", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			resp := `{"addPullRequestReviewThreadReply": {"comment": {"id": "PRRC_new", "url": "https://github.com/o/r/pull/1#discussion_r1"}}}`
			return json.Unmarshal([]byte(resp), response)
		})

		result, err := client.ReplyThread(ReplyThreadInput{ThreadID: "PRRT_1", Body: "Done in abc123"})
		if err != nil {
			t.Fatalf("ReplyThread() unexpected error: %v", err)
		}
		if result.ID != "PRRC_new" {
			t.Errorf("result.ID = %q, want %q", result.ID, "PRRC_new")
		}
		if result.URL != "https://github.com/o/r/pull/1#discussion_r1" {
			t.Errorf("result.URL = %q, want the discussion url", result.URL)
		}
	})

	t.Run("empty thread ID", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.ReplyThread(ReplyThreadInput{ThreadID: "", Body: "hi"})
		if err == nil {
			t.Error("ReplyThread() expected error for empty thread ID")
		}
	})

	t.Run("invalid thread ID format", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.ReplyThread(ReplyThreadInput{ThreadID: "PRRC_wrongkind", Body: "hi"})
		if err == nil {
			t.Error("ReplyThread() expected error for non-thread node ID")
		}
		if err != nil && !strings.Contains(err.Error(), "expected GraphQL thread node ID") {
			t.Errorf("error = %q, want containing 'expected GraphQL thread node ID'", err.Error())
		}
	})

	t.Run("empty body", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.ReplyThread(ReplyThreadInput{ThreadID: "PRRT_1", Body: "   "})
		if err == nil {
			t.Error("ReplyThread() expected error for empty body")
		}
	})

	t.Run("empty ID in response", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			return json.Unmarshal([]byte(`{"addPullRequestReviewThreadReply": {"comment": {"id": "", "url": ""}}}`), response)
		})
		_, err := client.ReplyThread(ReplyThreadInput{ThreadID: "PRRT_1", Body: "hi"})
		if err == nil {
			t.Error("ReplyThread() expected error for empty comment ID in response")
		}
	})

	t.Run("GraphQL error", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			return errors.New("mutation failed")
		})
		_, err := client.ReplyThread(ReplyThreadInput{ThreadID: "PRRT_1", Body: "hi"})
		if err == nil {
			t.Error("ReplyThread() expected error for GraphQL failure")
		}
	})
}

func TestClientResolveThread(t *testing.T) {
	t.Run("successful resolve", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			return json.Unmarshal([]byte(`{"resolveReviewThread": {"thread": {"id": "PRRT_1", "isResolved": true}}}`), response)
		})

		result, err := client.ResolveThread("PRRT_1")
		if err != nil {
			t.Fatalf("ResolveThread() unexpected error: %v", err)
		}
		if !result.IsResolved {
			t.Error("result.IsResolved = false, want true")
		}
		if result.ThreadID != "PRRT_1" {
			t.Errorf("result.ThreadID = %q, want %q", result.ThreadID, "PRRT_1")
		}
	})

	t.Run("empty thread ID", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.ResolveThread("")
		if err == nil {
			t.Error("ResolveThread() expected error for empty thread ID")
		}
	})

	t.Run("invalid thread ID format", func(t *testing.T) {
		client := newTestClient(nil)
		_, err := client.ResolveThread("PRRC_notathread")
		if err == nil {
			t.Error("ResolveThread() expected error for non-thread node ID")
		}
	})

	t.Run("GraphQL error", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			return errors.New("mutation failed")
		})
		_, err := client.ResolveThread("PRRT_1")
		if err == nil {
			t.Error("ResolveThread() expected error for GraphQL failure")
		}
	})
}

func TestClientThreadIDByComment(t *testing.T) {
	threadsResp := `{
		"repository": {
			"pullRequest": {
				"reviewThreads": {
					"totalCount": 2,
					"nodes": [
						{"id": "PRRT_a", "isResolved": false, "path": "a.go", "line": 1, "comments": {"nodes": [
							{"id": "PRRC_head_a", "body": "root", "author": {"login": "u"}, "pullRequestReview": {"state": "COMMENTED"}},
							{"id": "PRRC_reply_a", "body": "reply", "author": {"login": "v"}, "pullRequestReview": {"state": "COMMENTED"}}
						]}},
						{"id": "PRRT_b", "isResolved": false, "path": "b.go", "line": 2, "comments": {"nodes": [
							{"id": "PRRC_head_b", "body": "other", "author": {"login": "u"}, "pullRequestReview": {"state": "COMMENTED"}}
						]}}
					]
				}
			}
		}
	}`

	t.Run("matches head comment", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			return json.Unmarshal([]byte(threadsResp), response)
		})
		pr := &PRRef{Owner: "o", Repo: "r", Number: 1}
		id, err := client.ThreadIDByComment(pr, "PRRC_head_b")
		if err != nil {
			t.Fatalf("ThreadIDByComment() unexpected error: %v", err)
		}
		if id != "PRRT_b" {
			t.Errorf("threadID = %q, want %q", id, "PRRT_b")
		}
	})

	t.Run("matches reply comment", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			return json.Unmarshal([]byte(threadsResp), response)
		})
		pr := &PRRef{Owner: "o", Repo: "r", Number: 1}
		id, err := client.ThreadIDByComment(pr, "PRRC_reply_a")
		if err != nil {
			t.Fatalf("ThreadIDByComment() unexpected error: %v", err)
		}
		if id != "PRRT_a" {
			t.Errorf("threadID = %q, want %q", id, "PRRT_a")
		}
	})

	t.Run("no match", func(t *testing.T) {
		client := newTestClient(func(query string, variables map[string]interface{}, response interface{}) error {
			return json.Unmarshal([]byte(threadsResp), response)
		})
		pr := &PRRef{Owner: "o", Repo: "r", Number: 1}
		_, err := client.ThreadIDByComment(pr, "PRRC_missing")
		if err == nil {
			t.Error("ThreadIDByComment() expected error when no thread contains the comment")
		}
	})

	t.Run("empty comment ID", func(t *testing.T) {
		client := newTestClient(nil)
		pr := &PRRef{Owner: "o", Repo: "r", Number: 1}
		_, err := client.ThreadIDByComment(pr, "")
		if err == nil {
			t.Error("ThreadIDByComment() expected error for empty comment ID")
		}
	})
}
