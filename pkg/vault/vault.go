package vault

import (
	"log"

	"github.com/spf13/viper"
)

// vaultguardConfig is the struct containing vaultguard configuration
type vaultguardConfig struct {
	init     bool
	unseal   bool
	mount    bool
	policies bool
	gentoken bool
	Address  string `yaml:"listen_address" json:"listen_address"`
	Port     string `yaml:"listen_port" json:"listen_port"`
}

// vaultConfig is the struct containing vault configuration
type vaultConfig struct {
	vaultAddrs    []string
	vaultBackends []string
	vaultPolicies []string
}

// Entrypoint is the entrypoint to the vault pkg
func Entrypoint() error {

	// construct the vault configuration from viper config
	vgConf := viper.Sub("vaultguard")

	log.Printf("viper.Sub vaultguard config: %#v", vgConf)
	return nil
}
