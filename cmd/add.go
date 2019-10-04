package cmd

import (
	"eager/internal"
	"eager/pkg"
	"eager/pkg/cli"
	"eager/pkg/jira"
	"fmt"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addJiraCmd)

	addCmd.PersistentFlags().IntVar(&conf.Year, internal.FlagYear, time.Now().Year(), "specify the year")
	addCmd.PersistentFlags().IntVar(&conf.Month, internal.FlagMonth, int(time.Now().Month()), "specify the month")
	addCmd.PersistentFlags().IntVar(&conf.Day, internal.FlagDay, time.Now().Day(), "specify the day")
	addCmd.PersistentFlags().StringVar(&conf.Task, internal.FlagTask, "", "specify the task")
	addCmd.PersistentFlags().BoolVarP(&conf.Duration.Summarize, internal.FlagSummarize, "s", false, "sum effort on same day and task")
	addCmd.MarkFlagRequired(internal.FlagTask)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add worklog item",
	Long:  "Add a worklog item to the given store.",
	Args:  cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Call to root persistent pre run necessary. See https://github.com/spf13/cobra/issues/216
		err := cmd.Root().PersistentPreRunE(cmd, args)
		if err != nil {
			return err
		}
		return nil
	},
}

var addJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Add worklog item to Jira",
	Long:  "Add a worklog item to Atlassian Jira.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		duration, err := time.ParseDuration(args[0])
		if err != nil {
			return fmt.Errorf("not a valid duration '%s'", args[0])
		}
		jira.AddWorklogItem(
			pkg.NewHttpClient(),
			conf.Server(),
			conf.Userinfo(),
			conf.Year,
			time.Month(conf.Month),
			conf.Day,
			pkg.Task(conf.Task),
			duration,
			conf.Duration.Summarize,
			cli.Confirmation,
		)
		return nil
	},
}
