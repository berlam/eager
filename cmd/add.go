package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	// TODO mro Reactivate that command after implementing it
	//rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addJiraCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add worklog item",
	Long:  "Add a worklog item to your local store.",
	Args:  cobra.NoArgs,
}

var addJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Add worklog item to Jira",
	Long:  "Add a worklog item to Atlassian Jira.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
	},
}
