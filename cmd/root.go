package cmd

import (
	"eager/internal"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"os"
)

var cfgFile string

func init() {
	cobra.OnInitialize(func() {
		if cfgFile != "" {
			// Use Configuration file from the flag.
			viper.SetConfigFile(cfgFile)
		}
	})

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "specify the configuration file")
	rootCmd.PersistentFlags().BoolVarP(&internal.Config.Insecure, internal.FlagInsecure, "k", false, "use http instead of https")
	rootCmd.PersistentFlags().StringVarP(&internal.Config.Host, internal.FlagHost, "H", "", "specify the host to use for effort query")
	rootCmd.PersistentFlags().StringVarP(&internal.Config.Username, internal.FlagUsername, "u", "", "specify the username to use for server authentication")
	rootCmd.PersistentFlags().StringVarP(&internal.Config.Password, internal.FlagPassword, "p", "", "specify the password to use for server authentication")
	rootCmd.MarkPersistentFlagRequired(internal.FlagHost)
}

var rootCmd = &cobra.Command{
	Short:   "Eager is a tool to maintain a worklog",
	Long:    "Eager is a tool for real eager beavers to maintain and synchronize one worklog across different services.",
	Args:    cobra.NoArgs,
	Version: internal.Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cfgFile != "" {
			viper.BindPFlags(cmd.Flags())
			if err := viper.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return fmt.Errorf("can't read configuration. %s", err.Error())
				}
			}
			viper.Unmarshal(&internal.Config)
		}
		// Remove required annotation if the user has that flag given with viper.
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			annotations := flag.Annotations[cobra.BashCompOneRequiredFlag]
			if annotations != nil && annotations[0] == "true" && flag.Value.String() != flag.NoOptDefVal {
				delete(flag.Annotations, cobra.BashCompOneRequiredFlag)
			}
		})
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
