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
		var addr string
		var port string
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

		if _, ok := viper.Get("vaultguard.listen_address").(string); !ok {
			return errors.New("listen_address has to be of type string")
		}
		addr = viper.Get("vaultguard.listen_address").(string)
		if _, ok := viper.Get("vaultguard.listen_port").(string); !ok {
			return errors.New("listen_port has to be of type string")
		}
		port = viper.Get("vaultguard.listen_port").(string)
		sConf := listener.Config{
			Address: addr,
			Port:    port,
			Debug:   debugPtr,
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
var cfgFilePtr string

func init() {

	cobra.OnInitialize(initConfig)
	// flags available to all commands and subcommands
	RootCmd.PersistentFlags().BoolVarP(&debugPtr, "debug", "d", false, "turn on debug output")
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
