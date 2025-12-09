package output

import (
	"bytes"
	"testing"
)

func TestNewFormatter(t *testing.T) {
	var buf bytes.Buffer

	tests := []struct {
		name    string
		format  Format
		wantErr bool
	}{
		{
			name:    "table format",
			format:  FormatTable,
			wantErr: false,
		},
		{
			name:    "plain format",
			format:  FormatPlain,
			wantErr: false,
		},
		{
			name:    "json format",
			format:  FormatJSON,
			wantErr: false,
		},
		{
			name:    "unknown format",
			format:  Format("yaml"),
			wantErr: true,
		},
		{
			name:    "empty format",
			format:  Format(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewFormatter(tt.format, &buf)

			if tt.wantErr {
				if err == nil {
					t.Error("NewFormatter() expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("NewFormatter() unexpected error: %v", err)
				return
			}

			if formatter == nil {
				t.Error("NewFormatter() returned nil formatter")
			}
		})
	}
}

func TestFormatConstants(t *testing.T) {
	if FormatTable != "table" {
		t.Errorf("FormatTable = %q, want %q", FormatTable, "table")
	}
	if FormatPlain != "plain" {
		t.Errorf("FormatPlain = %q, want %q", FormatPlain, "plain")
	}
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", FormatJSON, "json")
	}
}

func TestResultTypes(t *testing.T) {
	tests := []struct {
		result   Result
		wantType string
	}{
		{CommentsResult{}, "comments"},
		{ViewResult{}, "view"},
		{AddResult{}, "add"},
		{EditResult{}, "edit"},
		{DeleteResult{}, "delete"},
		{SubmitResult{}, "submit"},
		{DiscardResult{}, "discard"},
		{NoOpResult{}, "noop"},
	}

	for _, tt := range tests {
		t.Run(tt.wantType, func(t *testing.T) {
			if got := tt.result.Type(); got != tt.wantType {
				t.Errorf("Type() = %q, want %q", got, tt.wantType)
			}
		})
	}
}

func TestCommentStruct(t *testing.T) {
	c := Comment{
		ID:     "PRRC_123",
		Path:   "main.go",
		Line:   42,
		Body:   "Fix this issue",
		State:  "pending",
		Author: "reviewer",
	}

	if c.ID != "PRRC_123" {
		t.Errorf("ID = %q, want %q", c.ID, "PRRC_123")
	}
	if c.Path != "main.go" {
		t.Errorf("Path = %q, want %q", c.Path, "main.go")
	}
	if c.Line != 42 {
		t.Errorf("Line = %d, want %d", c.Line, 42)
	}
	if c.Body != "Fix this issue" {
		t.Errorf("Body = %q, want %q", c.Body, "Fix this issue")
	}
	if c.State != "pending" {
		t.Errorf("State = %q, want %q", c.State, "pending")
	}
	if c.Author != "reviewer" {
		t.Errorf("Author = %q, want %q", c.Author, "reviewer")
	}
}

func TestViewThreadStruct(t *testing.T) {
	thread := ViewThread{
		ID:       "PRRT_123",
		Path:     "file.go",
		Line:     10,
		Resolved: false,
		Comments: []ViewThreadComment{
			{ID: "C1", Author: "user1", Body: "comment 1"},
			{ID: "C2", Author: "user2", Body: "comment 2"},
		},
	}

	if thread.ID != "PRRT_123" {
		t.Errorf("ID = %q, want %q", thread.ID, "PRRT_123")
	}
	if len(thread.Comments) != 2 {
		t.Errorf("Comments length = %d, want %d", len(thread.Comments), 2)
	}
	if thread.Resolved {
		t.Error("Resolved = true, want false")
	}
}

func TestCommentsResultStruct(t *testing.T) {
	result := CommentsResult{
		PRRef:      "owner/repo#123",
		IncludeIDs: true,
		Groups: []CommentGroup{
			{
				Author: "reviewer1",
				Comments: []*Comment{
					{ID: "C1", Body: "comment 1"},
				},
			},
		},
	}

	if result.PRRef != "owner/repo#123" {
		t.Errorf("PRRef = %q, want %q", result.PRRef, "owner/repo#123")
	}
	if !result.IncludeIDs {
		t.Error("IncludeIDs = false, want true")
	}
	if len(result.Groups) != 1 {
		t.Errorf("Groups length = %d, want %d", len(result.Groups), 1)
	}
}
