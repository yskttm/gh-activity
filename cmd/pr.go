package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yskttm/gh-activity/internal/github"
)

var prCmd = &cobra.Command{
	Use:   "pr <username>",
	Short: "List merged PRs by author",
	Args:  cobra.ExactArgs(1),
	RunE:  runPR,
}

func init() {
	today := time.Now().Format("2006-01-02")
	prCmd.Flags().String("from", today, "start date (YYYY-MM-DD)")
	prCmd.Flags().String("to", today, "end date (YYYY-MM-DD)")
	prCmd.Flags().StringSlice("repos", nil, "target repositories, comma-separated (e.g. org/repo1,org/repo2)")
	prCmd.Flags().StringSlice("fields", AllFieldNames, "fields to display, comma-separated")
}

func runPR(cmd *cobra.Command, args []string) error {
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

	fmt.Fprintf(os.Stderr, "Fetching PRs for %s (%s to %s)...\n", username, from, to)
	if len(repos) > 0 {
		fmt.Fprintf(os.Stderr, "Repos: %s\n", strings.Join(repos, " "))
	}
	fmt.Fprintln(os.Stderr)

	allItems, err := client.Search("author:"+username+" type:pr"+repoQuery, "merged", from, to)
	if err != nil {
		return err
	}

	github.SortItems(allItems)
	return printResults(allItems, fields, format)
}
