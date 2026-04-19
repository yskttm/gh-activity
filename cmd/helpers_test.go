package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yskttm/gh-activity/internal/github"
)

func TestFormatDate(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2024-01-15T10:30:00Z", "2024-01-15"},
		{"2024-12-31T23:59:59Z", "2024-12-31"},
		{"", ""},
		{"invalid", "invalid"}, // パース失敗時はそのまま返す
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := formatDate(tt.input); got != tt.want {
				t.Errorf("formatDate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateFields(t *testing.T) {
	tests := []struct {
		name    string
		fields  []string
		wantErr bool
	}{
		{"all valid", AllFieldNames, false},
		{"subset valid", []string{"no", "repo", "title"}, false},
		{"unknown field", []string{"no", "unknown"}, true},
		{"empty", []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFields(tt.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFields(%v) error = %v, wantErr %v", tt.fields, err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "unknown field") {
				t.Errorf("error message should contain 'unknown field', got: %v", err)
			}
		})
	}
}

func TestPrintCSV(t *testing.T) {
	items := []github.SearchItem{
		{
			Number:        42,
			Title:         "Fix bug",
			HTMLURL:       "https://github.com/org/repo/issues/42",
			RepositoryURL: "https://api.github.com/repos/org/repo",
		},
	}
	fields := []string{"no", "repo", "number", "title"}

	var out bytes.Buffer
	if err := printCSV(&out, items, fields); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 row), got %d", len(lines))
	}
	if lines[0] != "No.,Repo,Number,Title" {
		t.Errorf("unexpected header: %q", lines[0])
	}
	if !strings.Contains(lines[1], "Fix bug") {
		t.Errorf("row should contain title, got: %q", lines[1])
	}
	if !strings.Contains(lines[1], "org/repo") {
		t.Errorf("row should contain repo, got: %q", lines[1])
	}
}

func TestPrintResults_NoResults(t *testing.T) {
	var out bytes.Buffer
	if err := printResults(&out, []github.SearchItem{}, []string{"no", "title"}, "table"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "No results found.") {
		t.Errorf("expected 'No results found.', got: %q", out.String())
	}
}

func TestBuildRepoQuery(t *testing.T) {
	tests := []struct {
		name  string
		repos []string
		want  string
	}{
		{"no repos", []string{}, ""},
		{"single repo", []string{"org/repo1"}, " repo:org/repo1"},
		{"multiple repos", []string{"org/repo1", "org/repo2"}, " repo:org/repo1 repo:org/repo2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildRepoQuery(tt.repos); got != tt.want {
				t.Errorf("buildRepoQuery(%v) = %q, want %q", tt.repos, got, tt.want)
			}
		})
	}
}
