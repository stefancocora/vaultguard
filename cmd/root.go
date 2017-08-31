/*
Copyright 2015 All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"log"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefancocora/vaultguard/pkg/listener"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "vaultguard",
	Short: "vaultguard automates various aspects of vault",
	Long: `vaultguard automates various aspects of hashicorp vault
                by reading a config file and taking action
                on a cluster of vault servers.
                ...

                Documentation is available at http://doesnotexist.yet`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// required flags
		configFile := cmd.Flags().Lookup("config").Value.String()
		if debugPtr {
			// viper.Debug()
			log.Printf("config file being used: %v", viper.ConfigFileUsed())
			log.Printf("config content vaultguard: %v", viper.Get("vaultguard"))
			log.Printf("config content vault: %v", viper.Get("vault"))
		}
		if configFile == "" {
			return errors.New("unable to continue due to missing configuration")
		}

		sConf := listener.DbgConfig{
			Debug:       debugPtr,
			DebugConfig: debugConfPtr,
		}
		if err := listener.Entrypoint(sConf); err != nil {
			errm := fmt.Sprintf("error in the listener entrypoint: %v", err)
			return errors.Wrap(err, errm)
		}

		return nil
	},
}

//Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal("[FATAL] unable to add child commands to the root command")
	}
}

var debugPtr bool
var debugConfPtr bool
var cfgFilePtr string

func init() {

	cobra.OnInitialize(initConfig)
	// flags available to all commands and subcommands
	RootCmd.PersistentFlags().BoolVarP(&debugPtr, "debug", "d", false, "turn on debug output")
	RootCmd.PersistentFlags().BoolVarP(&debugConfPtr, "debugconfig", "", false, "turn on config struct debugging output")
	RootCmd.PersistentFlags().StringVar(&cfgFilePtr, "config", "", "config file (config will be searched in /vaultguard/config.yaml:/etc/vaultguard/config.yaml:$HOME/vaultguard/config.yaml)")

}

func initConfig() {
	// Don't forget to read config either from cfgFilePtr or from home directory!
	if cfgFilePtr != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFilePtr)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalf("[FATAL] unable to resolve home directory: %v", err)
		}

		viper.SetConfigType("yaml")

		// Search config in home directory with name ".cobra" (without extension).
		etcd := "/etc/vaultguard"
		viper.AddConfigPath(etcd)
		homed := fmt.Sprintf("%v/vaultguard", home)
		viper.AddConfigPath(homed)
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Can't read config: %v", err)
	}
}
