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

	showCmd.PersistentFlags().IntVar(&internal.Config.Year, internal.FlagYear, time.Now().Year(), "specify the year to query effort for")
	showCmd.PersistentFlags().IntVar(&internal.Config.Month, internal.FlagMonth, int(time.Now().Month()), "specify the month to query effort for")
	showCmd.PersistentFlags().BoolVarP(&internal.Config.Summarize, internal.FlagSummarize, "s", false, "summarize effort per day")
	showCmd.PersistentFlags().BoolVar(&internal.Config.PrintEmptyLine, internal.FlagPrintEmptyLine, false, "print empty line for missing day during summary")

	showBcsCmd.Flags().StringVar(&internal.Config.Report, internal.FlagReport, "", "specify the name of the report")
	showBcsCmd.Flags().StringArrayVar(&internal.Config.Projects, internal.FlagProjects, nil, "specify the oid of the project")
	showBcsCmd.MarkFlagRequired(internal.FlagReport)

	showJiraCmd.Flags().StringArrayVar(&internal.Config.Projects, internal.FlagProjects, nil, "specify the project key")
	showJiraCmd.Flags().StringArrayVar(&internal.Config.Users, internal.FlagUsers, nil, "show results for user")
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
		if !internal.Config.Summarize && internal.Config.PrintEmptyLine {
			return fmt.Errorf("empty lines (--%s) are only available for summaries (--%s)", internal.FlagPrintEmptyLine, internal.FlagSummarize)
		}
		if internal.Config.Projects != nil && internal.Config.PrintEmptyLine {
			return fmt.Errorf("empty lines (--%s) are not available for projects (--%s)", internal.FlagPrintEmptyLine, internal.FlagProjects)
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
		if internal.Config.Projects != nil && len(internal.Config.Projects) > 1 {
			return fmt.Errorf("only one project allowed")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if internal.Config.Projects == nil || len(internal.Config.Projects) == 0 {
			bcs.GetTimesheet(
				pkg.NewHttpClient(),
				internal.Config.Server(),
				internal.Config.Userinfo(),
				internal.Config.Year,
				time.Month(internal.Config.Month),
				internal.Config.Report,
			).Print(os.Stdout, false, internal.Config.Summarize, internal.Config.PrintEmptyLine)
		} else {
			bcs.GetBulkTimesheet(
				pkg.NewHttpClient(),
				internal.Config.Server(),
				internal.Config.Userinfo(),
				internal.Config.Year,
				time.Month(internal.Config.Month),
				pkg.Project(internal.Config.Projects[0]),
				internal.Config.Report,
			).Print(os.Stdout, true, internal.Config.Summarize, internal.Config.PrintEmptyLine)
		}
	},
}

var showJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Show worklog data from Jira",
	Long:  "Show your worklog data from Atlassian Jira.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if internal.Config.Users == nil || len(internal.Config.Users) == 0 {
			jira.GetTimesheet(
				pkg.NewHttpClient(),
				internal.Config.Server(),
				internal.Config.Userinfo(),
				internal.Config.Year,
				time.Month(internal.Config.Month),
				internal.Projects(internal.Config.Projects),
			).Print(os.Stdout, false, internal.Config.Summarize, internal.Config.PrintEmptyLine)
		} else {
			jira.GetBulkTimesheet(
				pkg.NewHttpClient(),
				internal.Config.Server(),
				internal.Config.Userinfo(),
				internal.Config.Year,
				time.Month(internal.Config.Month),
				internal.Projects(internal.Config.Projects),
				internal.Users(internal.Config.Users),
			).Print(os.Stdout, true, internal.Config.Summarize, internal.Config.PrintEmptyLine)
		}
	},
}
