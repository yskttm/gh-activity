package cmd

import (
	"bytes"
	"strings"
	"testing"

	gh "github.com/yskttm/gh-activity/internal/github"
)

func TestRunPR_NoResults(t *testing.T) {
	factory := newMockFactory([]mockResponse{
		{TotalCount: 0},
	})

	root := newTestRoot()
	root.AddCommand(newPRCmd(factory))

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"pr", "testuser", "--from=2024-01-01", "--to=2024-01-31"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "No results found.") {
		t.Errorf("expected 'No results found.', got: %q", out.String())
	}
}

func TestRunPR_WithResults(t *testing.T) {
	factory := newMockFactory([]mockResponse{
		{TotalCount: 2},
		{Items: []gh.SearchItem{
			{ID: 1, Number: 10, Title: "Add feature", RepositoryURL: "https://api.github.com/repos/org/repo"},
			{ID: 2, Number: 20, Title: "Fix bug", RepositoryURL: "https://api.github.com/repos/org/repo"},
		}},
	})

	root := newTestRoot()
	root.AddCommand(newPRCmd(factory))

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"pr", "testuser", "--from=2024-01-01", "--to=2024-01-31", "--fields=no,number,title"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outStr := out.String()
	if !strings.Contains(outStr, "Total: 2 item(s)") {
		t.Errorf("expected 2 items, got: %q", outStr)
	}
	if !strings.Contains(outStr, "Add feature") {
		t.Errorf("expected 'Add feature' in output, got: %q", outStr)
	}
}

func TestRunPR_CSVFormat(t *testing.T) {
	factory := newMockFactory([]mockResponse{
		{TotalCount: 1},
		{Items: []gh.SearchItem{
			{ID: 1, Number: 99, Title: "Refactor auth", RepositoryURL: "https://api.github.com/repos/org/repo"},
		}},
	})

	root := newTestRoot()
	root.AddCommand(newPRCmd(factory))

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"pr", "testuser", "--from=2024-01-01", "--to=2024-01-31", "--fields=no,number,title", "--format=csv"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 CSV lines, got %d: %q", len(lines), out.String())
	}
	if lines[0] != "No.,Number,Title" {
		t.Errorf("unexpected CSV header: %q", lines[0])
	}
	if !strings.Contains(lines[1], "Refactor auth") {
		t.Errorf("expected 'Refactor auth' in CSV row, got: %q", lines[1])
	}
}
