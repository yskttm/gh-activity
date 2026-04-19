package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yskttm/gh-activity/internal/github"
)

func newPRCmd(newClient func() (*github.Client, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr <username>",
		Short: "List merged PRs by author",
		Args:  cobra.ExactArgs(1),
	}

	today := time.Now().Format("2006-01-02")
	cmd.Flags().String("from", today, "start date (YYYY-MM-DD)")
	cmd.Flags().String("to", today, "end date (YYYY-MM-DD)")
	cmd.Flags().StringSlice("repos", nil, "target repositories, comma-separated (e.g. org/repo1,org/repo2)")
	cmd.Flags().StringSlice("fields", AllFieldNames, "fields to display, comma-separated")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runPR(cmd, args, newClient)
	}
	return cmd
}

func runPR(cmd *cobra.Command, args []string, newClient func() (*github.Client, error)) error {
	username := args[0]
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	repos, _ := cmd.Flags().GetStringSlice("repos")
	fields, _ := cmd.Flags().GetStringSlice("fields")
	if err := validateFields(fields); err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")

	client, err := newClient()
	if err != nil {
		return err
	}

	repoQuery := buildRepoQuery(repos)

	fmt.Fprintf(cmd.ErrOrStderr(), "Fetching PRs for %s (%s to %s)...\n", username, from, to)
	if len(repos) > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "Repos: %s\n", strings.Join(repos, " "))
	}
	fmt.Fprintln(cmd.ErrOrStderr())

	allItems, err := client.Search("author:"+username+" type:pr"+repoQuery, "merged", from, to)
	if err != nil {
		return err
	}

	github.SortItems(allItems)
	return printResults(cmd.OutOrStdout(), allItems, fields, format)
}
