package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yskttm/gh-activity/internal/github"
)

func newIssueCmd(newClient func() (*github.Client, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue <username>",
		Short: "List closed issues by assignee or author",
		Args:  cobra.ExactArgs(1),
	}

	today := time.Now().Format("2006-01-02")
	cmd.Flags().String("from", today, "start date (YYYY-MM-DD)")
	cmd.Flags().String("to", today, "end date (YYYY-MM-DD)")
	cmd.Flags().StringSlice("repos", nil, "target repositories, comma-separated (e.g. org/repo1,org/repo2)")
	cmd.Flags().StringSlice("fields", AllFieldNames, "fields to display, comma-separated")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runIssue(cmd, args, newClient)
	}
	return cmd
}

func runIssue(cmd *cobra.Command, args []string, newClient func() (*github.Client, error)) error {
	username := args[0]
	from, err := cmd.Flags().GetString("from")
	if err != nil {
		return err
	}
	to, err := cmd.Flags().GetString("to")
	if err != nil {
		return err
	}
	repos, err := cmd.Flags().GetStringSlice("repos")
	if err != nil {
		return err
	}
	fields, err := cmd.Flags().GetStringSlice("fields")
	if err != nil {
		return err
	}
	if err := validateFields(fields); err != nil {
		return err
	}
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	repoQuery := buildRepoQuery(repos)

	fmt.Fprintf(cmd.ErrOrStderr(), "Fetching issues for %s (%s to %s)...\n", username, from, to)
	if len(repos) > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "Repos: %s\n", strings.Join(repos, " "))
	}
	fmt.Fprintln(cmd.ErrOrStderr())

	fmt.Fprintln(cmd.ErrOrStderr(), "Fetching issues by assignee...")
	assigneeItems, err := client.Search("assignee:"+username+" type:issue"+repoQuery, "closed", from, to)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "Fetching issues by author...")
	authorItems, err := client.Search("author:"+username+" type:issue"+repoQuery, "closed", from, to)
	if err != nil {
		return err
	}

	allItems := github.UniqueByID(append(assigneeItems, authorItems...))
	github.SortItems(allItems)

	return printResults(cmd.OutOrStdout(), allItems, fields, format)
}
