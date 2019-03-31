package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	// TODO mro Reactivate that command after implementing it
	//rootCmd.AddCommand(removeCmd)
	removeCmd.AddCommand(removeJiraCmd)
}

var removeCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "Remove worklog item",
	Long:    "Remove a worklog item from your local store.",
	Args:    cobra.NoArgs,
}

var removeJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Remove worklog item from Jira",
	Long:  "Remove a worklog item from Atlassian Jira.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
	},
}
