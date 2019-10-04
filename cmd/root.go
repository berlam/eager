package cmd

import (
	"eager/internal"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
)

var conf internal.Configuration

func init() {
	cobra.OnInitialize(func() {
		conf, _ := rootCmd.PersistentFlags().GetString(internal.FlagConfiguration)
		if conf != "" {
			// Use Configuration file from the flag.
			viper.SetConfigFile(conf)
			return
		}

		// Lookup Configuration by going up the current directory
		wd, err := os.Getwd()
		if err != nil {
			return
		}
		for true {
			file := filepath.Clean(wd + string(filepath.Separator) + "eager.yaml")
			_, err := os.Stat(file)
			if os.IsNotExist(err) {
				nd, err := filepath.Abs(wd + string(filepath.Separator) + "..")
				if err != nil || nd == wd {
					break
				}
				wd = nd
				continue
			}
			// Found Configuration file
			viper.SetConfigFile(file)
			break
		}
	})

	rootCmd.PersistentFlags().StringP(internal.FlagConfiguration, "c", "", "specify the configuration file")
	rootCmd.PersistentFlags().BoolVarP(&conf.Http, internal.FlagHttp, "k", false, "use http instead of https")
	rootCmd.PersistentFlags().StringVarP(&conf.Host, internal.FlagHost, "H", "", "specify the host to use for effort query")
	rootCmd.PersistentFlags().StringVarP(&conf.Username, internal.FlagUsername, "u", "", "specify the username to use for server authentication")
	rootCmd.PersistentFlags().StringVarP(&conf.Password, internal.FlagPassword, "p", "", "specify the password to use for server authentication")
	rootCmd.MarkPersistentFlagRequired(internal.FlagHost)
}

var rootCmd = &cobra.Command{
	Short:   "Eager is a tool to maintain a worklog",
	Long:    "Eager is a tool for real eager beavers to maintain and synchronize one worklog across different services.",
	Args:    cobra.NoArgs,
	Version: internal.Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if viper.ConfigFileUsed() != "" {
			viper.BindPFlags(cmd.Flags())
			if err := viper.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return fmt.Errorf("can't read conf. %s", err.Error())
				}
			}
			viper.Unmarshal(&conf)
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
