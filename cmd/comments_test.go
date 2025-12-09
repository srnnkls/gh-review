package cmd

import (
	"testing"

	"github.com/srnnkls/gh-review/internal/output"
)

func TestMatchesFilters(t *testing.T) {
	// Save and restore package-level filter variables
	origStates := listStates
	origAuthor := listAuthor
	defer func() {
		listStates = origStates
		listAuthor = origAuthor
	}()

	tests := []struct {
		name       string
		comment    *output.Comment
		states     []string
		author     string
		wantMatch  bool
	}{
		{
			name:      "no filters matches all",
			comment:   &output.Comment{State: "pending", Author: "user1"},
			states:    nil,
			author:    "",
			wantMatch: true,
		},
		{
			name:      "state filter matches",
			comment:   &output.Comment{State: "pending", Author: "user1"},
			states:    []string{"pending"},
			author:    "",
			wantMatch: true,
		},
		{
			name:      "state filter case insensitive",
			comment:   &output.Comment{State: "PENDING", Author: "user1"},
			states:    []string{"pending"},
			author:    "",
			wantMatch: true,
		},
		{
			name:      "state filter no match",
			comment:   &output.Comment{State: "approved", Author: "user1"},
			states:    []string{"pending"},
			author:    "",
			wantMatch: false,
		},
		{
			name:      "multiple states - first matches",
			comment:   &output.Comment{State: "pending", Author: "user1"},
			states:    []string{"pending", "approved"},
			author:    "",
			wantMatch: true,
		},
		{
			name:      "multiple states - second matches",
			comment:   &output.Comment{State: "approved", Author: "user1"},
			states:    []string{"pending", "approved"},
			author:    "",
			wantMatch: true,
		},
		{
			name:      "multiple states - none match",
			comment:   &output.Comment{State: "changes_requested", Author: "user1"},
			states:    []string{"pending", "approved"},
			author:    "",
			wantMatch: false,
		},
		{
			name:      "author filter matches",
			comment:   &output.Comment{State: "pending", Author: "user1"},
			states:    nil,
			author:    "user1",
			wantMatch: true,
		},
		{
			name:      "author filter case insensitive",
			comment:   &output.Comment{State: "pending", Author: "USER1"},
			states:    nil,
			author:    "user1",
			wantMatch: true,
		},
		{
			name:      "author filter no match",
			comment:   &output.Comment{State: "pending", Author: "user2"},
			states:    nil,
			author:    "user1",
			wantMatch: false,
		},
		{
			name:      "both filters match",
			comment:   &output.Comment{State: "pending", Author: "user1"},
			states:    []string{"pending"},
			author:    "user1",
			wantMatch: true,
		},
		{
			name:      "state matches but author doesn't",
			comment:   &output.Comment{State: "pending", Author: "user2"},
			states:    []string{"pending"},
			author:    "user1",
			wantMatch: false,
		},
		{
			name:      "author matches but state doesn't",
			comment:   &output.Comment{State: "approved", Author: "user1"},
			states:    []string{"pending"},
			author:    "user1",
			wantMatch: false,
		},
		{
			name:      "empty state in comment",
			comment:   &output.Comment{State: "", Author: "user1"},
			states:    []string{"pending"},
			author:    "",
			wantMatch: false,
		},
		{
			name:      "empty author in comment with author filter",
			comment:   &output.Comment{State: "pending", Author: ""},
			states:    nil,
			author:    "user1",
			wantMatch: false,
		},
		{
			name:      "discussion state",
			comment:   &output.Comment{State: "discussion", Author: "user1"},
			states:    []string{"discussion"},
			author:    "",
			wantMatch: true,
		},
		{
			name:      "unresolved state",
			comment:   &output.Comment{State: "unresolved", Author: "user1"},
			states:    []string{"unresolved"},
			author:    "",
			wantMatch: true,
		},
		{
			name:      "changes_requested state",
			comment:   &output.Comment{State: "changes_requested", Author: "user1"},
			states:    []string{"changes_requested"},
			author:    "",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listStates = tt.states
			listAuthor = tt.author

			got := matchesFilters(tt.comment)
			if got != tt.wantMatch {
				t.Errorf("matchesFilters() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestMatchesFiltersWithEmptyComment(t *testing.T) {
	origStates := listStates
	origAuthor := listAuthor
	defer func() {
		listStates = origStates
		listAuthor = origAuthor
	}()

	listStates = nil
	listAuthor = ""

	// Empty comment with no filters should match
	comment := &output.Comment{}
	if !matchesFilters(comment) {
		t.Error("empty comment with no filters should match")
	}
}

func TestMatchesFiltersMultipleStatesOrLogic(t *testing.T) {
	origStates := listStates
	origAuthor := listAuthor
	defer func() {
		listStates = origStates
		listAuthor = origAuthor
	}()

	// Multiple states should use OR logic (match any)
	listStates = []string{"pending", "approved", "commented"}
	listAuthor = ""

	testCases := []struct {
		state string
		want  bool
	}{
		{"pending", true},
		{"approved", true},
		{"commented", true},
		{"changes_requested", false},
		{"dismissed", false},
	}

	for _, tc := range testCases {
		comment := &output.Comment{State: tc.state, Author: "user"}
		got := matchesFilters(comment)
		if got != tc.want {
			t.Errorf("state=%q: matchesFilters() = %v, want %v", tc.state, got, tc.want)
		}
	}
}
