package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/yskttm/gh-activity/internal/github"
)

var rootCmd = &cobra.Command{
	Use:   "gh-activity",
	Short: "GitHub Issue/PR activity aggregator",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("format", "table", "output format: table or csv")
	rootCmd.AddCommand(newIssueCmd(github.NewClient))
	rootCmd.AddCommand(newPRCmd(github.NewClient))
}
