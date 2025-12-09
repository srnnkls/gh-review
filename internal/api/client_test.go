package api

import (
	"testing"
)

func TestParsePRArg(t *testing.T) {
	tests := []struct {
		name         string
		arg          string
		wantNumber   int
		wantRepo     string
		wantErr      bool
		errContains  string
	}{
		{
			name:       "plain number",
			arg:        "123",
			wantNumber: 123,
			wantRepo:   "",
			wantErr:    false,
		},
		{
			name:       "hash prefix",
			arg:        "#456",
			wantNumber: 456,
			wantRepo:   "",
			wantErr:    false,
		},
		{
			name:       "number with whitespace",
			arg:        "  789  ",
			wantNumber: 789,
			wantRepo:   "",
			wantErr:    false,
		},
		{
			name:       "full GitHub URL",
			arg:        "https://github.com/owner/repo/pull/42",
			wantNumber: 42,
			wantRepo:   "owner/repo",
			wantErr:    false,
		},
		{
			name:       "GitHub URL without https",
			arg:        "github.com/myorg/myrepo/pull/100",
			wantNumber: 100,
			wantRepo:   "myorg/myrepo",
			wantErr:    false,
		},
		{
			name:       "GitHub URL with www",
			arg:        "https://www.github.com/test/project/pull/55",
			wantNumber: 55,
			wantRepo:   "test/project",
			wantErr:    false,
		},
		{
			name:       "http URL",
			arg:        "http://github.com/foo/bar/pull/1",
			wantNumber: 1,
			wantRepo:   "foo/bar",
			wantErr:    false,
		},
		{
			name:        "zero number",
			arg:         "0",
			wantErr:     true,
			errContains: "positive",
		},
		{
			name:        "negative number",
			arg:         "-5",
			wantErr:     true,
			errContains: "positive",
		},
		{
			name:        "non-numeric string",
			arg:         "abc",
			wantErr:     true,
			errContains: "invalid PR reference",
		},
		{
			name:        "empty string",
			arg:         "",
			wantErr:     true,
			errContains: "invalid PR reference",
		},
		{
			name:        "whitespace only",
			arg:         "   ",
			wantErr:     true,
			errContains: "invalid PR reference",
		},
		{
			name:        "hash only",
			arg:         "#",
			wantErr:     true,
			errContains: "invalid PR reference",
		},
		{
			name:       "large number",
			arg:        "99999",
			wantNumber: 99999,
			wantRepo:   "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			number, repo, err := ParsePRArg(tt.arg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePRArg(%q) expected error, got nil", tt.arg)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ParsePRArg(%q) error = %q, want containing %q", tt.arg, err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParsePRArg(%q) unexpected error: %v", tt.arg, err)
				return
			}

			if number != tt.wantNumber {
				t.Errorf("ParsePRArg(%q) number = %d, want %d", tt.arg, number, tt.wantNumber)
			}
			if repo != tt.wantRepo {
				t.Errorf("ParsePRArg(%q) repo = %q, want %q", tt.arg, repo, tt.wantRepo)
			}
		})
	}
}

func TestPRRefString(t *testing.T) {
	tests := []struct {
		name   string
		pr     PRRef
		want   string
	}{
		{
			name:   "standard ref",
			pr:     PRRef{Owner: "owner", Repo: "repo", Number: 123},
			want:   "owner/repo#123",
		},
		{
			name:   "with org name",
			pr:     PRRef{Owner: "my-org", Repo: "my-repo", Number: 1},
			want:   "my-org/my-repo#1",
		},
		{
			name:   "large number",
			pr:     PRRef{Owner: "a", Repo: "b", Number: 99999},
			want:   "a/b#99999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pr.String()
			if got != tt.want {
				t.Errorf("PRRef.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewPRRefValidation(t *testing.T) {
	// Test that NewPRRef validates number
	_, err := NewPRRef(0, "owner/repo")
	if err == nil {
		t.Error("NewPRRef(0, ...) expected error for zero number")
	}

	_, err = NewPRRef(-1, "owner/repo")
	if err == nil {
		t.Error("NewPRRef(-1, ...) expected error for negative number")
	}
}

// contains checks if s contains substr (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
