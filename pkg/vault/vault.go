package vault

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// GuardConfig is the struct containing vaultguard configuration
type GuardConfig struct {
	init     bool
	unseal   bool
	mount    bool
	policies bool
	gentoken bool
	Address  string `yaml:"listen_address" json:"listen_address"`
	Port     string `yaml:"listen_port" json:"listen_port"`
}

// Config is the struct containing vault configuration
type Config struct {
	vaultAddrs    map[string][]string
	vaultBackends map[string]string
	vaultPolicies []string
}

// New sets up the Config struct from the configuration file.
// It reads from the configuration all the vaultguard config options.
func (g *GuardConfig) New() error {

	// construct the vault configuration from viper config
	vgConf := viper.Sub("vaultguard")

	if err := vgConf.Unmarshal(g); err != nil {
		errm := fmt.Sprintf("unable to unmarshal vaultguard subconfig: %v", err)
		return errors.New(errm)
	}
	return nil

}

// RunInit func
func RunInit(ctx context.Context, vgc GuardConfig, wg *sync.WaitGroup, retErrCh chan error) error {

	defer wg.Done()

	// fake work
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// timeout := time.After(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			log.Println("vaultInit: caller has asked us to stop processing work; shutting down.")
			return nil
		// case <-timeout:
		// 	errm := fmt.Sprintf("vaultInit: timeout while receiving response from the vault server")
		// 	return errors.New(errm)
		case t := <-ticker.C:
			// do some work
			fmt.Printf("vaultInit: working - %s\n", t.UTC().Format("20060102-150405.000000000"))
		}
	}
}

// RunUnseal is unsealing the vault
func RunUnseal(ctx context.Context, vgc GuardConfig, wg *sync.WaitGroup, retErrCh chan error) error {

	defer wg.Done()

	// fake work
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// timeout := time.After(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			log.Println("vaultUnseal: caller has asked us to stop processing work; shutting down.")
			return nil
		// case <-timeout:
		// 	errm := fmt.Sprintf("vaultUnseal: timeout while receiving response from the vault server")
		// 	return errors.New(errm)
		case t := <-ticker.C:
			// do some work
			fmt.Printf("vaultUnseal: working - %s\n", t.UTC().Format("20060102-150405.000000000"))
		}
	}
}
