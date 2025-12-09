package cmd

import (
	"testing"

	"github.com/srnnkls/gh-review/internal/output"
)

func TestOutputFormat(t *testing.T) {
	// Save and restore the global flag
	origFlag := formatFlag
	defer func() { formatFlag = origFlag }()

	tests := []struct {
		flag string
		want output.Format
	}{
		{"table", output.FormatTable},
		{"plain", output.FormatPlain},
		{"json", output.FormatJSON},
		{"", output.FormatTable},      // default
		{"unknown", output.FormatTable}, // unknown defaults to table
		{"TABLE", output.FormatTable},   // case sensitive - doesn't match "table"
		{"JSON", output.FormatTable},    // case sensitive - doesn't match "json"
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			formatFlag = tt.flag
			got := outputFormat()
			if got != tt.want {
				t.Errorf("outputFormat() with flag=%q = %q, want %q", tt.flag, got, tt.want)
			}
		})
	}
}

func TestResolvePRWithURL(t *testing.T) {
	// Save and restore
	origRepo := repoFlag
	defer func() { repoFlag = origRepo }()

	// Test that URL takes precedence over -R flag
	repoFlag = "other/repo"

	pr, err := resolvePR("https://github.com/owner/repo/pull/123")
	if err != nil {
		t.Fatalf("resolvePR() error: %v", err)
	}

	if pr.Owner != "owner" {
		t.Errorf("Owner = %q, want %q", pr.Owner, "owner")
	}
	if pr.Repo != "repo" {
		t.Errorf("Repo = %q, want %q", pr.Repo, "repo")
	}
	if pr.Number != 123 {
		t.Errorf("Number = %d, want %d", pr.Number, 123)
	}
}

func TestResolvePRInvalidArg(t *testing.T) {
	origRepo := repoFlag
	defer func() { repoFlag = origRepo }()

	tests := []struct {
		name string
		arg  string
	}{
		{"empty", ""},
		{"non-numeric", "abc"},
		{"zero", "0"},
		{"negative", "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoFlag = "owner/repo" // valid repo
			_, err := resolvePR(tt.arg)
			if err == nil {
				t.Errorf("resolvePR(%q) expected error", tt.arg)
			}
		})
	}
}

func TestResolvePRURLVariants(t *testing.T) {
	origRepo := repoFlag
	defer func() { repoFlag = origRepo }()

	tests := []struct {
		name   string
		arg    string
		owner  string
		repo   string
		number int
	}{
		{
			name:   "https URL",
			arg:    "https://github.com/owner/repo/pull/42",
			owner:  "owner",
			repo:   "repo",
			number: 42,
		},
		{
			name:   "http URL",
			arg:    "http://github.com/org/project/pull/1",
			owner:  "org",
			repo:   "project",
			number: 1,
		},
		{
			name:   "www URL",
			arg:    "https://www.github.com/test/lib/pull/99",
			owner:  "test",
			repo:   "lib",
			number: 99,
		},
		{
			name:   "no protocol",
			arg:    "github.com/foo/bar/pull/55",
			owner:  "foo",
			repo:   "bar",
			number: 55,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoFlag = "" // no override

			pr, err := resolvePR(tt.arg)
			if err != nil {
				t.Fatalf("resolvePR(%q) error: %v", tt.arg, err)
			}

			if pr.Owner != tt.owner {
				t.Errorf("Owner = %q, want %q", pr.Owner, tt.owner)
			}
			if pr.Repo != tt.repo {
				t.Errorf("Repo = %q, want %q", pr.Repo, tt.repo)
			}
			if pr.Number != tt.number {
				t.Errorf("Number = %d, want %d", pr.Number, tt.number)
			}
		})
	}
}

func TestRootCmdFlags(t *testing.T) {
	// Test that flags are registered
	formatFlagDef := rootCmd.PersistentFlags().Lookup("format")
	if formatFlagDef == nil {
		t.Error("format flag not registered")
	}
	if formatFlagDef.Shorthand != "f" {
		t.Errorf("format flag shorthand = %q, want %q", formatFlagDef.Shorthand, "f")
	}
	if formatFlagDef.DefValue != "table" {
		t.Errorf("format flag default = %q, want %q", formatFlagDef.DefValue, "table")
	}

	repoFlagDef := rootCmd.PersistentFlags().Lookup("repo")
	if repoFlagDef == nil {
		t.Error("repo flag not registered")
	}
	if repoFlagDef.Shorthand != "R" {
		t.Errorf("repo flag shorthand = %q, want %q", repoFlagDef.Shorthand, "R")
	}
}

func TestRootCmdBasic(t *testing.T) {
	if rootCmd.Use != "gh-review" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "gh-review")
	}
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}
}
