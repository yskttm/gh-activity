package cmd

import (
	"os"

	"github.com/spf13/cobra"
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
	rootCmd.AddCommand(issueCmd)
	rootCmd.AddCommand(prCmd)
}
