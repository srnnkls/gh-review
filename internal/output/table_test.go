package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTruncateBody(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string unchanged",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length unchanged",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "long string truncated",
			input:  "hello world",
			maxLen: 8,
			want:   "hello wo...",
		},
		{
			name:   "newlines replaced",
			input:  "hello\nworld",
			maxLen: 20,
			want:   "hello world",
		},
		{
			name:   "newlines and truncation",
			input:  "hello\nworld\nfoo",
			maxLen: 10,
			want:   "hello worl...",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "multiple newlines",
			input:  "a\nb\nc\nd",
			maxLen: 20,
			want:   "a b c d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateBody(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateBody(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestTableFormatterCommentsResult(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := CommentsResult{
		PRRef: "owner/repo#123",
		Groups: []CommentGroup{
			{
				Author: "reviewer",
				Comments: []*Comment{
					{Path: "main.go", Line: 10, Body: "Fix this", State: "pending"},
				},
			},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	// Check that output contains expected elements
	if !strings.Contains(output, "@reviewer") {
		t.Error("output should contain author header")
	}
	if !strings.Contains(output, "pending") {
		t.Error("output should contain state")
	}
	if !strings.Contains(output, "main.go") {
		t.Error("output should contain path")
	}
}

func TestTableFormatterCommentsResultFlatMode(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := CommentsResult{
		PRRef: "owner/repo#123",
		Groups: []CommentGroup{
			{
				Author: "", // flat mode - no author grouping
				Comments: []*Comment{
					{Path: "file.go", Line: 1, Body: "comment", State: "approved", Author: "user1"},
				},
			},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	// In flat mode, author should appear per row
	if !strings.Contains(output, "user1") {
		t.Error("output should contain author per row in flat mode")
	}
}

func TestTableFormatterCommentsResultWithIDs(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := CommentsResult{
		PRRef:      "owner/repo#123",
		IncludeIDs: true,
		Groups: []CommentGroup{
			{
				Author: "user",
				Comments: []*Comment{
					{ID: "PRRC_123", Path: "file.go", Line: 1, Body: "comment", State: "pending"},
				},
			},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "PRRC_123") {
		t.Error("output should contain comment ID when IncludeIDs is true")
	}
}

func TestTableFormatterViewResult(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := ViewResult{
		PRRef: "owner/repo#456",
		Threads: []ViewThread{
			{
				ID:       "PRRT_1",
				Path:     "main.go",
				Line:     20,
				Resolved: false,
				Comments: []ViewThreadComment{
					{ID: "C1", Author: "reviewer", Body: "Please fix this issue"},
				},
			},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "unresolved") {
		t.Error("output should contain 'unresolved' status")
	}
	if !strings.Contains(output, "main.go:20") {
		t.Error("output should contain path:line")
	}
	if !strings.Contains(output, "@reviewer") {
		t.Error("output should contain comment author")
	}
}

func TestTableFormatterViewResultResolved(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := ViewResult{
		PRRef: "owner/repo#456",
		Threads: []ViewThread{
			{
				Path:     "file.go",
				Line:     10,
				Resolved: true,
				Comments: []ViewThreadComment{
					{Author: "user", Body: "Done"},
				},
			},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "resolved") {
		t.Error("output should contain 'resolved' status")
	}
}

func TestTableFormatterViewResultWithIDs(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := ViewResult{
		PRRef:      "owner/repo#456",
		IncludeIDs: true,
		Threads: []ViewThread{
			{
				ID:   "PRRT_123",
				Path: "file.go",
				Line: 10,
				Comments: []ViewThreadComment{
					{ID: "PRRC_456", Author: "user", Body: "comment"},
				},
			},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "PRRT_123") {
		t.Error("output should contain thread ID when IncludeIDs is true")
	}
	if !strings.Contains(output, "PRRC_456") {
		t.Error("output should contain comment ID when IncludeIDs is true")
	}
}

func TestTableFormatterAddResult(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := AddResult{Path: "main.go", Line: 42}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Added") {
		t.Error("output should contain 'Added'")
	}
	if !strings.Contains(output, "main.go:42") {
		t.Error("output should contain path:line")
	}
}

func TestTableFormatterEditResult(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := EditResult{CommentID: "PRRC_123"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Updated") {
		t.Error("output should contain 'Updated'")
	}
	if !strings.Contains(output, "PRRC_123") {
		t.Error("output should contain comment ID")
	}
}

func TestTableFormatterDeleteResult(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := DeleteResult{CommentID: "PRRC_456"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Deleted") {
		t.Error("output should contain 'Deleted'")
	}
	if !strings.Contains(output, "PRRC_456") {
		t.Error("output should contain comment ID")
	}
}

func TestTableFormatterSubmitResult(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := SubmitResult{Verdict: "APPROVE"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Submitted") {
		t.Error("output should contain 'Submitted'")
	}
	if !strings.Contains(output, "APPROVE") {
		t.Error("output should contain verdict")
	}
}

func TestTableFormatterDiscardResult(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := DiscardResult{ReviewID: "PRR_789"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Discarded") {
		t.Error("output should contain 'Discarded'")
	}
	if !strings.Contains(output, "PRR_789") {
		t.Error("output should contain review ID")
	}
}

func TestTableFormatterNoOpResult(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := NoOpResult{Message: "Nothing to do"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Nothing to do") {
		t.Error("output should contain message")
	}
}

func TestTableFormatterMultipleGroups(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

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

	output := buf.String()

	if !strings.Contains(output, "@user1") {
		t.Error("output should contain first author")
	}
	if !strings.Contains(output, "@user2") {
		t.Error("output should contain second author")
	}
}

func TestTableFormatterLongBodyTruncation(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	longBody := strings.Repeat("x", 100)
	result := CommentsResult{
		PRRef: "owner/repo#1",
		Groups: []CommentGroup{
			{Author: "user", Comments: []*Comment{{Body: longBody, State: "pending"}}},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	// Body should be truncated and contain "..."
	if !strings.Contains(output, "...") {
		t.Error("long body should be truncated with '...'")
	}
}

func TestTableFormatterPathTruncation(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	longPath := "very/long/path/to/some/deeply/nested/file.go"
	result := CommentsResult{
		PRRef: "owner/repo#1",
		Groups: []CommentGroup{
			{Author: "user", Comments: []*Comment{{Path: longPath, Line: 1, Body: "c", State: "pending"}}},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	// Path should be truncated with "..." prefix
	if !strings.Contains(output, "...") {
		t.Error("long path should be truncated")
	}
}

func TestTableFormatterGlobalComment(t *testing.T) {
	var buf bytes.Buffer
	formatter := newTableFormatter(&buf)

	result := CommentsResult{
		PRRef: "owner/repo#1",
		Groups: []CommentGroup{
			{Author: "user", Comments: []*Comment{{Path: "", Line: 0, Body: "Global comment", State: "discussion"}}},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "(global)") {
		t.Error("comment without path should show '(global)'")
	}
}
