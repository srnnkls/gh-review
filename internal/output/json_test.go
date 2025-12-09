package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJSONFormatterCommentsResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatJSON, &buf)

	result := CommentsResult{
		PRRef:      "owner/repo#123",
		IncludeIDs: false,
		Groups: []CommentGroup{
			{
				Author: "reviewer",
				Comments: []*Comment{
					{ID: "PRRC_1", Path: "main.go", Line: 10, Body: "Fix this", State: "pending", Author: "reviewer"},
				},
			},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if parsed["pr"] != "owner/repo#123" {
		t.Errorf("pr = %v, want %v", parsed["pr"], "owner/repo#123")
	}

	groups, ok := parsed["groups"].([]interface{})
	if !ok || len(groups) != 1 {
		t.Fatalf("groups not found or wrong length")
	}

	group := groups[0].(map[string]interface{})
	if group["author"] != "reviewer" {
		t.Errorf("author = %v, want %v", group["author"], "reviewer")
	}

	comments := group["comments"].([]interface{})
	comment := comments[0].(map[string]interface{})

	// ID should not be included when IncludeIDs is false
	if _, hasID := comment["id"]; hasID && comment["id"] != "" {
		t.Error("ID should not be included when IncludeIDs is false")
	}

	if comment["state"] != "pending" {
		t.Errorf("state = %v, want %v", comment["state"], "pending")
	}
}

func TestJSONFormatterCommentsResultWithIDs(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatJSON, &buf)

	result := CommentsResult{
		PRRef:      "owner/repo#123",
		IncludeIDs: true,
		Groups: []CommentGroup{
			{
				Author: "reviewer",
				Comments: []*Comment{
					{ID: "PRRC_1", Body: "comment", State: "approved"},
				},
			},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(buf.Bytes(), &parsed)

	groups := parsed["groups"].([]interface{})
	comments := groups[0].(map[string]interface{})["comments"].([]interface{})
	comment := comments[0].(map[string]interface{})

	if comment["id"] != "PRRC_1" {
		t.Errorf("id = %v, want %v", comment["id"], "PRRC_1")
	}
}

func TestJSONFormatterViewResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatJSON, &buf)

	result := ViewResult{
		PRRef:      "owner/repo#456",
		IncludeIDs: true,
		Threads: []ViewThread{
			{
				ID:       "PRRT_1",
				Path:     "file.go",
				Line:     20,
				Resolved: false,
				Comments: []ViewThreadComment{
					{ID: "C1", Author: "user1", Body: "comment body"},
				},
			},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(buf.Bytes(), &parsed)

	if parsed["pr"] != "owner/repo#456" {
		t.Errorf("pr = %v, want %v", parsed["pr"], "owner/repo#456")
	}

	threads := parsed["threads"].([]interface{})
	if len(threads) != 1 {
		t.Fatalf("threads length = %d, want 1", len(threads))
	}

	thread := threads[0].(map[string]interface{})
	if thread["id"] != "PRRT_1" {
		t.Errorf("thread id = %v, want %v", thread["id"], "PRRT_1")
	}
	if thread["path"] != "file.go" {
		t.Errorf("path = %v, want %v", thread["path"], "file.go")
	}
	if thread["resolved"] != false {
		t.Errorf("resolved = %v, want false", thread["resolved"])
	}
}

func TestJSONFormatterAddResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatJSON, &buf)

	result := AddResult{Path: "main.go", Line: 42}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(buf.Bytes(), &parsed)

	if parsed["action"] != "added" {
		t.Errorf("action = %v, want %v", parsed["action"], "added")
	}
	if parsed["path"] != "main.go" {
		t.Errorf("path = %v, want %v", parsed["path"], "main.go")
	}
	if parsed["line"] != float64(42) {
		t.Errorf("line = %v, want %v", parsed["line"], 42)
	}
}

func TestJSONFormatterEditResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatJSON, &buf)

	result := EditResult{CommentID: "PRRC_123"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(buf.Bytes(), &parsed)

	if parsed["action"] != "edited" {
		t.Errorf("action = %v, want %v", parsed["action"], "edited")
	}
	if parsed["comment_id"] != "PRRC_123" {
		t.Errorf("comment_id = %v, want %v", parsed["comment_id"], "PRRC_123")
	}
}

func TestJSONFormatterDeleteResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatJSON, &buf)

	result := DeleteResult{CommentID: "PRRC_456"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(buf.Bytes(), &parsed)

	if parsed["action"] != "deleted" {
		t.Errorf("action = %v, want %v", parsed["action"], "deleted")
	}
	if parsed["comment_id"] != "PRRC_456" {
		t.Errorf("comment_id = %v, want %v", parsed["comment_id"], "PRRC_456")
	}
}

func TestJSONFormatterSubmitResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatJSON, &buf)

	result := SubmitResult{Verdict: "APPROVE"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(buf.Bytes(), &parsed)

	if parsed["action"] != "submitted" {
		t.Errorf("action = %v, want %v", parsed["action"], "submitted")
	}
	if parsed["verdict"] != "APPROVE" {
		t.Errorf("verdict = %v, want %v", parsed["verdict"], "APPROVE")
	}
}

func TestJSONFormatterDiscardResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatJSON, &buf)

	result := DiscardResult{ReviewID: "PRR_789"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(buf.Bytes(), &parsed)

	if parsed["action"] != "discarded" {
		t.Errorf("action = %v, want %v", parsed["action"], "discarded")
	}
	if parsed["review_id"] != "PRR_789" {
		t.Errorf("review_id = %v, want %v", parsed["review_id"], "PRR_789")
	}
}

func TestJSONFormatterNoOpResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatJSON, &buf)

	result := NoOpResult{Message: "Nothing to do"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(buf.Bytes(), &parsed)

	if parsed["action"] != "noop" {
		t.Errorf("action = %v, want %v", parsed["action"], "noop")
	}
	if parsed["message"] != "Nothing to do" {
		t.Errorf("message = %v, want %v", parsed["message"], "Nothing to do")
	}
}

func TestJSONFormatterEmptyGroups(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatJSON, &buf)

	result := CommentsResult{
		PRRef:  "owner/repo#1",
		Groups: []CommentGroup{},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(buf.Bytes(), &parsed)

	groups := parsed["groups"].([]interface{})
	if len(groups) != 0 {
		t.Errorf("groups length = %d, want 0", len(groups))
	}
}

func TestJSONFormatterMultipleGroups(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatJSON, &buf)

	result := CommentsResult{
		PRRef: "owner/repo#1",
		Groups: []CommentGroup{
			{Author: "user1", Comments: []*Comment{{Body: "c1", State: "pending"}}},
			{Author: "user2", Comments: []*Comment{{Body: "c2", State: "approved"}}},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(buf.Bytes(), &parsed)

	groups := parsed["groups"].([]interface{})
	if len(groups) != 2 {
		t.Errorf("groups length = %d, want 2", len(groups))
	}
}

func TestJSONOutputIsValidJSON(t *testing.T) {
	results := []Result{
		CommentsResult{PRRef: "o/r#1", Groups: []CommentGroup{}},
		ViewResult{PRRef: "o/r#2", Threads: []ViewThread{}},
		AddResult{Path: "f.go", Line: 1},
		EditResult{CommentID: "C1"},
		DeleteResult{CommentID: "C2"},
		SubmitResult{Verdict: "APPROVE"},
		DiscardResult{ReviewID: "R1"},
		NoOpResult{Message: "done"},
	}

	for _, result := range results {
		var buf bytes.Buffer
		formatter, _ := NewFormatter(FormatJSON, &buf)

		err := formatter.Format(result)
		if err != nil {
			t.Errorf("Format(%T) error: %v", result, err)
			continue
		}

		var parsed interface{}
		if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
			t.Errorf("Format(%T) produced invalid JSON: %v\nOutput: %s", result, err, buf.String())
		}
	}
}
