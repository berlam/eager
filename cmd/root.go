package cmd

import (
	"bytes"
	"eager/internal"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
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
		file, err := os.Getwd()
		if err != nil {
			return
		}
		viper.SetConfigName("eager")
		for true {
			viper.AddConfigPath(file)
			nd, err := filepath.Abs(file + string(filepath.Separator) + "..")
			if err != nil || nd == file {
				break
			}
			file = nd
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
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			// Unbind configuration flag, otherwise argument would overwrite config file values
			if flag.Name == internal.FlagConfiguration {
				return
			}
			// Bind the rest
			err := viper.BindPFlag(flag.Name, flag)
			if err != nil {
				return
			}
		})
		err := viper.ReadInConfig()
		if viper.ConfigFileUsed() != "" {
			if err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return fmt.Errorf("cannot read conf. %s", err.Error())
				}
			}
			if viper.GetString(internal.FlagConfiguration) != "" {
				known := make(map[string]*struct{})
				err := parseConfiguration(known, viper.ConfigFileUsed())
				if err != nil {
					return fmt.Errorf("cannot read conf. %s", err.Error())
				}
			}
			err = viper.Unmarshal(&conf)
			if err != nil {
				return err
			}
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

func parseConfiguration(configurations map[string]*struct{}, parent string) error {
	parent, _ = filepath.Abs(parent)
	if configurations[parent] != nil {
		return fmt.Errorf("there is a recursion in you configuration definition")
	}
	configurations[parent] = &struct{}{}
	file, err := ioutil.ReadFile(parent)
	if err != nil {
		return fmt.Errorf("cannot read file %s", parent)
	}
	err = viper.ReadConfig(bytes.NewReader(file))
	if err != nil {
		return fmt.Errorf("cannot parse file %s", parent)
	}
	parent = viper.GetString(internal.FlagConfiguration)
	if parent == "" {
		// Nu further parent. Take that already read config
		return nil
	}
	err = parseConfiguration(configurations, parent)
	if err != nil {
		return err
	}
	// Merge the rest
	return viper.MergeConfig(bytes.NewReader(file))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
