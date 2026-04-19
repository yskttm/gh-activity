package cmd

import (
	"bytes"
	"strings"
	"testing"

	gh "github.com/yskttm/gh-activity/internal/github"
)

func TestRunIssue_NoResults(t *testing.T) {
	factory := newMockFactory([]mockResponse{
		{TotalCount: 0}, // assignee checkCount → 0 items
		{TotalCount: 0}, // author checkCount → 0 items
	})

	root := newTestRoot()
	root.AddCommand(newIssueCmd(factory))

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"issue", "testuser", "--from=2024-01-01", "--to=2024-01-31"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "No results found.") {
		t.Errorf("expected 'No results found.', got: %q", out.String())
	}
}

func TestRunIssue_MergesAssigneeAndAuthor(t *testing.T) {
	// assignee: items 1,2 / author: items 2,3 → unique: 1,2,3
	factory := newMockFactory([]mockResponse{
		{TotalCount: 2}, // assignee checkCount
		{Items: []gh.SearchItem{
			{ID: 1, Number: 1, RepositoryURL: "https://api.github.com/repos/org/repo"},
			{ID: 2, Number: 2, RepositoryURL: "https://api.github.com/repos/org/repo"},
		}},
		{TotalCount: 2}, // author checkCount
		{Items: []gh.SearchItem{
			{ID: 2, Number: 2, RepositoryURL: "https://api.github.com/repos/org/repo"},
			{ID: 3, Number: 3, RepositoryURL: "https://api.github.com/repos/org/repo"},
		}},
	})

	root := newTestRoot()
	root.AddCommand(newIssueCmd(factory))

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"issue", "testuser", "--from=2024-01-01", "--to=2024-01-31", "--fields=no,number"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Total: 3 item(s)") {
		t.Errorf("expected 3 unique items, got: %q", out.String())
	}
}

func TestRunIssue_CSVFormat(t *testing.T) {
	factory := newMockFactory([]mockResponse{
		{TotalCount: 1}, // assignee checkCount
		{Items: []gh.SearchItem{
			{ID: 1, Number: 42, Title: "Fix bug", RepositoryURL: "https://api.github.com/repos/org/repo"},
		}},
		{TotalCount: 0}, // author checkCount → 0 items
	})

	root := newTestRoot()
	root.AddCommand(newIssueCmd(factory))

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"issue", "testuser", "--from=2024-01-01", "--to=2024-01-31", "--fields=no,number,title", "--format=csv"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 CSV lines (header + 1 row), got %d: %q", len(lines), out.String())
	}
	if lines[0] != "No.,Number,Title" {
		t.Errorf("unexpected CSV header: %q", lines[0])
	}
	if !strings.Contains(lines[1], "Fix bug") {
		t.Errorf("expected 'Fix bug' in CSV row, got: %q", lines[1])
	}
}

func TestRunIssue_InvalidField(t *testing.T) {
	factory := newMockFactory(nil)

	root := newTestRoot()
	root.AddCommand(newIssueCmd(factory))
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"issue", "testuser", "--fields=no,nonexistent"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid field, got nil")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("expected 'unknown field' in error, got: %v", err)
	}
}
