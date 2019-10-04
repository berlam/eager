package cmd

import (
	"eager/internal"
	"eager/pkg"
	"eager/pkg/cli"
	"eager/pkg/jira"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.AddCommand(removeJiraCmd)

	removeCmd.PersistentFlags().IntVar(&conf.Year, internal.FlagYear, time.Now().Year(), "specify the year")
	removeCmd.PersistentFlags().IntVar(&conf.Month, internal.FlagMonth, int(time.Now().Month()), "specify the month")
	removeCmd.PersistentFlags().IntVar(&conf.Day, internal.FlagDay, time.Now().Day(), "specify the day")
	removeCmd.PersistentFlags().StringVar(&conf.Task, internal.FlagTask, "", "specify the task")
	removeCmd.MarkFlagRequired(internal.FlagTask)
}

var removeCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "Remove worklog item",
	Long:    "Remove a worklog item from your local store.",
	Args:    cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Call to root persistent pre run necessary. See https://github.com/spf13/cobra/issues/216
		err := cmd.Root().PersistentPreRunE(cmd, args)
		if err != nil {
			return err
		}
		return nil
	},
}

var removeJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Remove worklog item from Jira",
	Long:  "Remove a worklog item from Atlassian Jira.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		jira.RemoveWorklogItem(
			pkg.NewHttpClient(),
			conf.Server(),
			conf.Userinfo(),
			conf.Year,
			time.Month(conf.Month),
			conf.Day,
			pkg.Task(conf.Task),
			cli.Confirmation,
		)
	},
}
