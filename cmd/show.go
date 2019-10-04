package cmd

import (
	"eager/internal"
	"eager/pkg"
	"eager/pkg/bcs"
	"eager/pkg/jira"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"time"
)

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.AddCommand(showBcsCmd, showJiraCmd)

	showCmd.PersistentFlags().IntVar(&conf.Year, internal.FlagYear, time.Now().Year(), "specify the year to query effort for")
	showCmd.PersistentFlags().IntVar(&conf.Month, internal.FlagMonth, int(time.Now().Month()), "specify the month to query effort for")
	showCmd.PersistentFlags().BoolVarP(&conf.Duration.Summarize, internal.FlagSummarize, "s", false, "summarize effort per day")
	showCmd.PersistentFlags().BoolVar(&conf.Duration.Empty, internal.FlagEmpty, false, "print empty durations during summary")
	showCmd.PersistentFlags().BoolVar(&conf.Duration.Decimal, internal.FlagDecimal, false, "display duration as decimal hour")
	showCmd.PersistentFlags().BoolVar(&conf.Duration.Negate, internal.FlagNegate, false, "negate decimal duration")

	showBcsCmd.Flags().StringVar(&conf.Report, internal.FlagReport, "", "specify the name of the report")
	showBcsCmd.Flags().StringArrayVar(&conf.Projects, internal.FlagProjects, nil, "specify the oid of the project")
	showBcsCmd.MarkFlagRequired(internal.FlagReport)

	showJiraCmd.Flags().StringArrayVar(&conf.Projects, internal.FlagProjects, nil, "specify the project key")
	showJiraCmd.Flags().StringArrayVar(&conf.Users, internal.FlagUsers, nil, "show results for user (or user=id, where id is the account id)")
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show worklog",
	Long:  "Show worklog from the given store.",
	Args:  cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Call to root persistent pre run necessary. See https://github.com/spf13/cobra/issues/216
		err := cmd.Root().PersistentPreRunE(cmd, args)
		if err != nil {
			return err
		}
		if !conf.Duration.Summarize && conf.Duration.Empty {
			return fmt.Errorf("empty durations (--%s) are only available for summaries (--%s)", internal.FlagEmpty, internal.FlagSummarize)
		}
		if !conf.Duration.Decimal && conf.Duration.Negate {
			return fmt.Errorf("negative durations (--%s) are only available for decimal values (--%s)", internal.FlagNegate, internal.FlagDecimal)
		}
		return nil
	},
}

var showBcsCmd = &cobra.Command{
	Use:   "bcs",
	Short: "Show worklog from BCS",
	Long:  "Show your worklog data from Projektron BCS.",
	Args:  cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := cmd.Parent().PersistentPreRunE(cmd, args)
		if err != nil {
			return err
		}
		if conf.Projects != nil && len(conf.Projects) > 1 {
			return fmt.Errorf("only one project allowed")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if conf.Projects == nil || len(conf.Projects) == 0 {
			bcs.GetTimesheet(
				pkg.NewHttpClient(),
				conf.Server(),
				conf.Userinfo(),
				conf.Year,
				time.Month(conf.Month),
				conf.Report,
			).Print(os.Stdout, false, &conf.Duration)
		} else {
			bcs.GetBulkTimesheet(
				pkg.NewHttpClient(),
				conf.Server(),
				conf.Userinfo(),
				conf.Year,
				time.Month(conf.Month),
				pkg.Projects(conf.Projects),
				conf.Report,
			).Print(os.Stdout, true, &conf.Duration)
		}
	},
}

var showJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Show worklog from Jira",
	Long:  "Show your worklog data from Atlassian Jira.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if conf.Users == nil || len(conf.Users) == 0 {
			jira.GetTimesheet(
				pkg.NewHttpClient(),
				conf.Server(),
				conf.Userinfo(),
				conf.Year,
				time.Month(conf.Month),
				pkg.Projects(conf.Projects),
			).Print(os.Stdout, false, &conf.Duration)
		} else {
			jira.GetBulkTimesheet(
				pkg.NewHttpClient(),
				conf.Server(),
				conf.Userinfo(),
				conf.Year,
				time.Month(conf.Month),
				pkg.Projects(conf.Projects),
				pkg.Users(conf.Users),
			).Print(os.Stdout, true, &conf.Duration)
		}
	},
}
