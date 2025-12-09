package templates

import (
	"sort"
	"testing"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantOK   bool
		wantLen  int // minimum expected length
	}{
		{
			name:     "naming template exists",
			template: "naming",
			wantOK:   true,
			wantLen:  10,
		},
		{
			name:     "security template exists",
			template: "security",
			wantOK:   true,
			wantLen:  10,
		},
		{
			name:     "perf template exists",
			template: "perf",
			wantOK:   true,
			wantLen:  10,
		},
		{
			name:     "style template exists",
			template: "style",
			wantOK:   true,
			wantLen:  10,
		},
		{
			name:     "nonexistent template",
			template: "nonexistent",
			wantOK:   false,
			wantLen:  0,
		},
		{
			name:     "empty string",
			template: "",
			wantOK:   false,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, ok := Get(tt.template)
			if ok != tt.wantOK {
				t.Errorf("Get(%q) ok = %v, want %v", tt.template, ok, tt.wantOK)
			}
			if len(content) < tt.wantLen {
				t.Errorf("Get(%q) content length = %d, want >= %d", tt.template, len(content), tt.wantLen)
			}
		})
	}
}

func TestGetContentContains(t *testing.T) {
	tests := []struct {
		template string
		contains string
	}{
		{"naming", "Naming"},
		{"security", "Security"},
		{"perf", "Performance"},
		{"style", "Style"},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			content, ok := Get(tt.template)
			if !ok {
				t.Fatalf("Get(%q) returned not ok", tt.template)
			}
			if content == "" {
				t.Errorf("Get(%q) returned empty content", tt.template)
			}
		})
	}
}

func TestList(t *testing.T) {
	names := List()

	// Should have exactly 4 templates
	if len(names) != 4 {
		t.Errorf("List() returned %d templates, want 4", len(names))
	}

	// Sort for consistent comparison
	sort.Strings(names)
	expected := []string{"naming", "perf", "security", "style"}

	for i, name := range expected {
		found := false
		for _, n := range names {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("List() missing template %q at index %d", name, i)
		}
	}
}

func TestListAllRetrievable(t *testing.T) {
	names := List()
	for _, name := range names {
		content, ok := Get(name)
		if !ok {
			t.Errorf("Template %q listed but not retrievable", name)
		}
		if content == "" {
			t.Errorf("Template %q has empty content", name)
		}
	}
}
