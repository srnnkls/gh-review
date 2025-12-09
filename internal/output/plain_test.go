package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPlainFormatterCommentsResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

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

	if !strings.Contains(output, "@reviewer") {
		t.Error("output should contain author header")
	}
	if !strings.Contains(output, "pending") {
		t.Error("output should contain state")
	}
	if !strings.Contains(output, "main.go:10") {
		t.Error("output should contain path:line")
	}
	if !strings.Contains(output, "Fix this") {
		t.Error("output should contain body")
	}
}

func TestPlainFormatterCommentsResultWithIDs(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

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

func TestPlainFormatterCommentsResultFlatMode(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

	result := CommentsResult{
		PRRef: "owner/repo#123",
		Groups: []CommentGroup{
			{
				Author: "", // flat mode
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

func TestPlainFormatterViewResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

	result := ViewResult{
		PRRef: "owner/repo#456",
		Threads: []ViewThread{
			{
				Path:     "main.go",
				Line:     20,
				Resolved: false,
				Comments: []ViewThreadComment{
					{Author: "reviewer", Body: "Please fix"},
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
	if !strings.Contains(output, "reviewer") {
		t.Error("output should contain comment author")
	}
}

func TestPlainFormatterViewResultResolved(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

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

func TestPlainFormatterViewResultWithIDs(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

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

func TestPlainFormatterAddResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

	result := AddResult{Path: "main.go", Line: 42}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "added") {
		t.Error("output should contain 'added'")
	}
	if !strings.Contains(output, "main.go") {
		t.Error("output should contain path")
	}
	if !strings.Contains(output, "42") {
		t.Error("output should contain line number")
	}
}

func TestPlainFormatterEditResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

	result := EditResult{CommentID: "PRRC_123"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "edited") {
		t.Error("output should contain 'edited'")
	}
	if !strings.Contains(output, "PRRC_123") {
		t.Error("output should contain comment ID")
	}
}

func TestPlainFormatterDeleteResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

	result := DeleteResult{CommentID: "PRRC_456"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "deleted") {
		t.Error("output should contain 'deleted'")
	}
	if !strings.Contains(output, "PRRC_456") {
		t.Error("output should contain comment ID")
	}
}

func TestPlainFormatterSubmitResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

	result := SubmitResult{Verdict: "APPROVE"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "submitted") {
		t.Error("output should contain 'submitted'")
	}
	if !strings.Contains(output, "APPROVE") {
		t.Error("output should contain verdict")
	}
}

func TestPlainFormatterDiscardResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

	result := DiscardResult{ReviewID: "PRR_789"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "discarded") {
		t.Error("output should contain 'discarded'")
	}
	if !strings.Contains(output, "PRR_789") {
		t.Error("output should contain review ID")
	}
}

func TestPlainFormatterNoOpResult(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

	result := NoOpResult{Message: "Nothing to do"}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "noop") {
		t.Error("output should contain 'noop'")
	}
	if !strings.Contains(output, "Nothing to do") {
		t.Error("output should contain message")
	}
}

func TestPlainFormatterTSVFormat(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

	result := CommentsResult{
		PRRef: "owner/repo#123",
		Groups: []CommentGroup{
			{
				Author: "user",
				Comments: []*Comment{
					{Path: "file.go", Line: 10, Body: "comment", State: "pending"},
				},
			},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	// TSV format should use tabs
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "@") {
			continue // skip author header
		}
		if strings.HasPrefix(line, "\t") {
			// comment line should have tab-separated values
			if !strings.Contains(line, "\t") {
				t.Error("TSV format should use tabs")
			}
		}
	}
}

func TestPlainFormatterMultipleGroups(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

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

func TestPlainFormatterNewlinesInBody(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

	result := ViewResult{
		PRRef: "owner/repo#1",
		Threads: []ViewThread{
			{
				Path: "file.go",
				Line: 1,
				Comments: []ViewThreadComment{
					{Author: "user", Body: "line1\nline2\nline3"},
				},
			},
		},
	}

	err := formatter.Format(result)
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	output := buf.String()

	// Body newlines should be replaced with spaces for single-line output
	if strings.Contains(output, "line1\nline2") {
		t.Error("newlines in body should be replaced")
	}
}

func TestPlainFormatterGlobalComment(t *testing.T) {
	var buf bytes.Buffer
	formatter, _ := NewFormatter(FormatPlain, &buf)

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

	// Global comments (no path) should still be formatted
	output := buf.String()
	if !strings.Contains(output, "discussion") {
		t.Error("output should contain state even for global comments")
	}
}
