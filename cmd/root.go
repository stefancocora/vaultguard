package cmd

import (
	"fmt"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "vaultguard",
	Short: "vaultguard automates various aspects of vault",
	Long: `vaultguard automates various aspects of vault
                by reading a config file and taking action
                on a cluster of vault servers.
                ...

                Documentation is available at http://doesnotexist.yet`,
	Run: func(cmd *cobra.Command, args []string) {},
}

//Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal("[FATAL] unable to add child commands to the root command")
	}
}

var debugPtr bool
var cfgFilePtr string

func init() {

	cobra.OnInitialize(initConfig)
	// flags available to all commands and subcommands
	RootCmd.PersistentFlags().BoolVarP(&debugPtr, "debug", "d", false, "turn on debug output")
	RootCmd.PersistentFlags().StringVar(&cfgFilePtr, "config", "/etc/vaultguard/config.yaml", "config file (config will be searched in /etc/vaultguard/config.yaml:$HOME/vaultguard/config.yaml)")

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
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cobra" (without extension).
		homed := fmt.Sprintf("%v/vaultguard", home)
		viper.AddConfigPath(homed)
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}

	if debugPtr {
		log.Printf("config file being used: %v", viper.ConfigFileUsed())
		log.Printf("config content vaultguard: %v", viper.Get("vaultguard"))
		log.Printf("config content vault: %v", viper.Get("vault"))
	}
}
