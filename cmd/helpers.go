package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/yskttm/gh-activity/internal/github"
)

var AllFieldNames = []string{
	"no", "repo", "number", "url", "title", "state",
	"creator", "assignee", "labels", "milestone", "comments",
	"created_at", "closed_at", "merged_at", "draft",
}

type fieldDef struct {
	header string
	value  func(i int, item github.SearchItem) string
}

var fieldDefs = map[string]fieldDef{
	"no": {"No.", func(i int, _ github.SearchItem) string {
		return fmt.Sprintf("%d", i+1)
	}},
	"repo": {"Repo", func(_ int, item github.SearchItem) string {
		parts := strings.Split(item.RepositoryURL, "/")
		return strings.Join(parts[len(parts)-2:], "/")
	}},
	"number": {"Number", func(_ int, item github.SearchItem) string {
		return fmt.Sprintf("%d", item.Number)
	}},
	"url":   {"URL", func(_ int, item github.SearchItem) string { return item.HTMLURL }},
	"title": {"Title", func(_ int, item github.SearchItem) string { return item.Title }},
	"state": {"State", func(_ int, item github.SearchItem) string { return item.State }},
	"creator": {"Creator", func(_ int, item github.SearchItem) string {
		return item.User.Login
	}},
	"assignee": {"Assignee", func(_ int, item github.SearchItem) string {
		logins := make([]string, len(item.Assignees))
		for j, a := range item.Assignees {
			logins[j] = a.Login
		}
		return strings.Join(logins, ", ")
	}},
	"labels": {"Labels", func(_ int, item github.SearchItem) string {
		names := make([]string, len(item.Labels))
		for j, l := range item.Labels {
			names[j] = l.Name
		}
		return strings.Join(names, ", ")
	}},
	"milestone": {"Milestone", func(_ int, item github.SearchItem) string {
		if item.Milestone != nil {
			return item.Milestone.Title
		}
		return ""
	}},
	"comments": {"Comments", func(_ int, item github.SearchItem) string {
		return fmt.Sprintf("%d", item.Comments)
	}},
	"created_at": {"Created At", func(_ int, item github.SearchItem) string {
		return formatDate(item.CreatedAt)
	}},
	"closed_at": {"Closed At", func(_ int, item github.SearchItem) string {
		return formatDate(item.ClosedAt)
	}},
	"merged_at": {"Merged At", func(_ int, item github.SearchItem) string {
		if item.PullRequest != nil {
			return formatDate(item.PullRequest.MergedAt)
		}
		return ""
	}},
	"draft": {"Draft", func(_ int, item github.SearchItem) string {
		if item.Draft {
			return "yes"
		}
		return ""
	}},
}

func formatDate(s string) string {
	if s == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("2006-01-02")
}

func buildRepoQuery(repos []string) string {
	var sb strings.Builder
	for _, r := range repos {
		sb.WriteString(" repo:")
		sb.WriteString(r)
	}
	return sb.String()
}

func validateFields(fields []string) error {
	for _, f := range fields {
		if _, ok := fieldDefs[f]; !ok {
			return fmt.Errorf("unknown field %q (available: %s)", f, strings.Join(AllFieldNames, ", "))
		}
	}
	return nil
}

func printResults(items []github.SearchItem, fields []string, format string) error {
	if len(items) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	switch format {
	case "csv":
		return printCSV(items, fields)
	default:
		return printTable(items, fields)
	}
}

func printTable(items []github.SearchItem, fields []string) error {
	fmt.Printf("Total: %d item(s)\n\n", len(items))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	headers := make([]string, len(fields))
	for i, f := range fields {
		headers[i] = fieldDefs[f].header
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for i, item := range items {
		values := make([]string, len(fields))
		for j, f := range fields {
			values[j] = fieldDefs[f].value(i, item)
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}

	return w.Flush()
}

func printCSV(items []github.SearchItem, fields []string) error {
	w := csv.NewWriter(os.Stdout)

	headers := make([]string, len(fields))
	for i, f := range fields {
		headers[i] = fieldDefs[f].header
	}
	if err := w.Write(headers); err != nil {
		return err
	}

	for i, item := range items {
		values := make([]string, len(fields))
		for j, f := range fields {
			values[j] = fieldDefs[f].value(i, item)
		}
		if err := w.Write(values); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}
