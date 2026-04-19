package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yskttm/gh-activity/internal/github"
)

var issueCmd = &cobra.Command{
	Use:   "issue <username>",
	Short: "List closed issues by assignee or author",
	Args:  cobra.ExactArgs(1),
	RunE:  runIssue,
}

func init() {
	today := time.Now().Format("2006-01-02")
	issueCmd.Flags().String("from", today, "start date (YYYY-MM-DD)")
	issueCmd.Flags().String("to", today, "end date (YYYY-MM-DD)")
	issueCmd.Flags().StringSlice("repos", nil, "target repositories, comma-separated (e.g. org/repo1,org/repo2)")
	issueCmd.Flags().StringSlice("fields", AllFieldNames, "fields to display, comma-separated")
}

func runIssue(cmd *cobra.Command, args []string) error {
	username := args[0]
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	repos, _ := cmd.Flags().GetStringSlice("repos")
	fields, _ := cmd.Flags().GetStringSlice("fields")
	if err := validateFields(fields); err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")

	client, err := github.NewClient()
	if err != nil {
		return err
	}

	repoQuery := buildRepoQuery(repos)

	fmt.Fprintf(os.Stderr, "Fetching issues for %s (%s to %s)...\n", username, from, to)
	if len(repos) > 0 {
		fmt.Fprintf(os.Stderr, "Repos: %s\n", strings.Join(repos, " "))
	}
	fmt.Fprintln(os.Stderr)

	fmt.Fprintln(os.Stderr, "Fetching issues by assignee...")
	assigneeItems, err := client.Search("assignee:"+username+" type:issue"+repoQuery, "closed", from, to)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Fetching issues by author...")
	authorItems, err := client.Search("author:"+username+" type:issue"+repoQuery, "closed", from, to)
	if err != nil {
		return err
	}

	allItems := github.UniqueByID(append(assigneeItems, authorItems...))
	github.SortItems(allItems)

	return printResults(allItems, fields, format)
}
