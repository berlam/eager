package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	// TODO mro Reactivate that command after implementing it
	//rootCmd.AddCommand(pushCmd)
	pushCmd.AddCommand(pushJiraCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push worklog",
	Long:  "Push your local worklog to a remote store.",
	Args:  cobra.NoArgs,
}

var pushJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Push worklog to Jira",
	Long:  "Push your local worklog to Atlassian Jira.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
	},
}
