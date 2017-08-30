package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

var dbgVaultPkg bool
var dbgVaultConf bool

// Config is a vaultguard top level config
type Config struct {
	App `yaml:"app" json:"app"`
}

// App is the vaultguard application config made up of the vault config and the vaultguard config
type App struct {
	VConf       `yaml:"vault" json:"vault"`
	GuardConfig `yaml:"vaultguard" json:"vaultguard"`
}

// VConf holds the vault configuration definition
// VConf is the struct containing the various vault related config options
type VConf struct {
	Backends  []Backend   `yaml:"vault_backends" json:"vault_backends"`
	Endpoints []Endpoints `yaml:"vault_endpoints" json:"vault_endpoints"`
	Policy    []Policy    `yaml:"vault_policies" json:"vault_policies"`
}

// GuardConfig is the struct containing vaultguard configuration
type GuardConfig struct {
	Init     bool   `yaml:"init" json:"init"`
	Unseal   bool   `yaml:"unseal" json:"unseal"`
	Mount    bool   `yaml:"mount" json:"mount"`
	Policies bool   `yaml:"policies" json:"policies"`
	Gentoken bool   `yaml:"gentoken" json:"gentoken"`
	Address  string `yaml:"listen_address" json:"listen_address"`
	Port     string `yaml:"listen_port" json:"listen_port"`
}

// Endpoints holds the config for how to get to vault cluster endpoints
type Endpoints struct {
	Type  string `yaml:"type" json:"type"`
	Specs []Spec `yaml:"spec" json:"spec"`
}

// Spec contains the overall Endpoint definition
type Spec struct {
	// ecs
	Cluster string `yaml:"cluster,omitempty" json:"cluster,omitempty"`
	Region  string `yaml:"region,omitempty" json:"region,omitempty"`
	// url
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
	// k8s
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Service   string `yaml:"service,omitempty" json:"service,omitempty"`
}

// Backend is a definition of a vault secret backend
type Backend struct {
	Type        string `yaml:"type" json:"type"`
	Mountpath   string `yaml:"mountpath" json:"mountpath"`
	Description string `yaml:"description" json:"description"`
}

// Policy is a definition of a policy
type Policy struct {
	Name   string `yaml:"name" json:"name"`
	Policy string `yaml:"policy" json:"policy"`
}

//
// EcsSpec is the Endpoint that holds the definition of the requirements to get to a vault service running in AWS ECS
type EcsSpec struct {
	Cluster string `yaml:"cluster" json:"cluster"`
	Region  string `yaml:"region" json:"region"`
}

// URLSpec is the Endpoint that  holds the definition of the requirements to get to a vault service running at a defined URL
type URLSpec struct {
	URL string `yaml:"url" json:"url"`
}

// K8sSpec is the Endpoint that  holds the definition of the requirements to get to a vault service running in a kubernetes cluster
type K8sSpec struct {
	Namespace string `yaml:"namespace" json:"namespace"`
	Service   string `yaml:"service" json:"service"`
}

// New sets up the Config struct from the configuration file.
// It reads from the configuration all the vaultguard config options.
func (g *Config) New() error {

	f, errr := ioutil.ReadFile(viper.ConfigFileUsed())
	if errr != nil {
		return errr
	}

	format := strings.TrimPrefix(filepath.Ext(viper.ConfigFileUsed()), ".")
	if dbgVaultConf {
		log.Printf("format of config file: %s", format)
	}

	switch format {
	case "json":
		decj := json.NewDecoder(bytes.NewReader(f))
		if err := decj.Decode(&g); err != nil {
			errm := fmt.Sprintf("error decoding config file: %v", err)
			return errors.New(errm)
		}
	case "yml":
		fallthrough
	case "yaml":
		if err := yaml.Unmarshal(f, g); err != nil {
			errm := fmt.Sprintf("unable to unmarshal yaml config file: %v", err)
			return errors.New(errm)
		}
	default:
		errm := fmt.Sprintf("unsupported config file format: %s", format)
		return errors.New(errm)
	}

	if dbgVaultConf {
		log.Println("entire config decoded")
		spew.Dump(g)
	}

	return nil

}

// WorkerID is used to assign goroutine workers a notion of identity, useful when logging
type WorkerID struct {
	Name string
	Type string
	ID   int
}

// RunInit func
func RunInit(ctx context.Context, vgc Config, wg *sync.WaitGroup, retErrCh chan error, dvCh chan map[string][]string, id WorkerID) error {

	defer wg.Done()
	defer log.Printf("%v%v: worker shutdown complete", id.Name, id.ID)

	dv := make(map[string][]string)
	dv = <-dvCh
	log.Printf("%v%v: received discovered vault endpoints: %v", id.Name, id.ID, dv)

	// fake work
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// timeout := time.After(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			log.Printf("%v%v: caller has asked us to stop processing work; shutting down.", id.Name, id.ID)
			return nil
		// case <-timeout:
		// 	errm := fmt.Sprintf("%v: timeout while receiving response from the vault server")
		// 	return errors.New(errm)
		case t := <-ticker.C:
			// do some work
			fmt.Printf("%v%v: working - %s\n", id.Name, id.ID, t.UTC().Format("20060102-150405.000000000"))
		}
	}
}

// RunUnseal is unsealing the vault
func RunUnseal(ctx context.Context, vgc Config, wg *sync.WaitGroup, retErrCh chan error, id WorkerID) error {

	defer wg.Done()
	defer log.Printf("%v%v: worker shutdown complete", id.Name, id.ID)

	// fake work
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// timeout := time.After(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			log.Printf("%v%v: caller has asked us to stop processing work; shutting down.", id.Name, id.ID)
			return nil
		// case <-timeout:
		// 	errm := fmt.Sprintf("%v: timeout while receiving response from the vault server")
		// 	return errors.New(errm)
		case t := <-ticker.C:
			// do some work
			fmt.Printf("%v%v: working - %s\n", id.Name, id.ID, t.UTC().Format("20060102-150405.000000000"))
		}
	}
}

// PropagateDebug propagates the debug flag from main into this pkg, when explicitly called
func PropagateDebug(dbg bool, confDbg bool) {
	dbgVaultPkg = dbg
	dbgVaultConf = confDbg
}
