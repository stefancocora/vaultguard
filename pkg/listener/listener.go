package listener

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	ecs "github.com/stefancocora/vaultguard/pkg/discover/aws"
	"github.com/stefancocora/vaultguard/pkg/server"
	vaultg "github.com/stefancocora/vaultguard/pkg/vault"
)

var debugListenerPtr bool
var debugListenerConf bool

// DbgConfig is the type that is used to pass configuration to the http server
type DbgConfig struct {
	Debug       bool
	DebugConfig bool
}

// workerID is used to assign goroutine workers a notion of identity, useful when logging
type workerID struct {
	Name string
	Type string
	ID   int
}

// Entrypoint represents the entrypoint in the server package
func Entrypoint(srvConfig DbgConfig) error {
	debugListenerPtr = srvConfig.Debug
	debugListenerConf = srvConfig.DebugConfig

	if debugListenerPtr {
		log.Printf("received config: %#v", srvConfig)
	}

	// step: read config file
	log.Println("reading config file")
	var vgconf vaultg.Config
	if err := vgconf.New(); err != nil {
		log.Printf("unable to create vaultguard configuration %v", err)
	}

	// step: launch long running daemon and additional workers
	run(srvConfig, vgconf)

	return nil
}

// run starts all long running threads and communication channels
func run(srvConfig DbgConfig, vgconf vaultg.Config) {

	defer log.Println("run: shutdown complete")

	// step: create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}

	// step: setup OS signal handler
	if debugListenerPtr {
		log.Println("run: listening for OS signals")
	}
	osStopCh := make(chan os.Signal, 1)
	signal.Notify(osStopCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	// step: start the HTTP server
	log.Println("run: starting the HTTPSrv")
	wg.Add(1)
	id := workerID{
		Name: "httpSrvWrk",
		Type: "HTTPSrv",
		ID:   1,
	}
	go runHTTPSrv(ctx, srvConfig, vgconf, wg, id)

	// step: discover vault servers
	// channel for discovered vault endpoints to send to init
	var dvinitCh = make(chan map[string][]string, 1)
	// channel for errors that we get during init phase
	retErrChInit := make(chan error)

	log.Println("run: starting the ecsDsc worker")
	dv := runEcsDsc(srvConfig, vgconf)
	dvinitCh <- dv

	// step: start vaultInit worker
	if debugListenerPtr {
		vaultg.PropagateDebug(srvConfig.Debug, srvConfig.DebugConfig)
	}
	if vgconf.GuardConfig.Init {
		log.Println("run: starting the vaultInit worker")
		wg.Add(1)

		id := vaultg.WorkerID{
			Name: "vaultInitWrk",
			Type: "init",
			ID:   1,
		}
		go vaultg.RunInit(ctx, vgconf, wg, retErrChInit, dvinitCh, id) // start vault Init worker
	} else {
		log.Printf("run: init phase is disabled in the config file: %v", vgconf.GuardConfig.Init)
	}

	// step: start vaultUnseal worker
	if debugListenerPtr {
		vaultg.PropagateDebug(srvConfig.Debug, srvConfig.DebugConfig)
	}
	retErrChUnseal := make(chan error)
	if vgconf.GuardConfig.Init {
		log.Println("run: starting the vaultUnseal worker")
		wg.Add(1)
		id := vaultg.WorkerID{
			Name: "vaultUnsealWrk",
			Type: "unseal",
			ID:   1,
		}
		go vaultg.RunUnseal(ctx, vgconf, wg, retErrChUnseal, id) // start vault Init worker
	} else {
		log.Printf("run: unseal phase is disabled in the config file: %v", vgconf.GuardConfig.Init)
	}

	// step: long running process
listenerloop:
	for {
		select {
		case sig := <-osStopCh:
			log.Printf("run: received termination signal ( %v ), asking all goroutines to stop.", sig)
			// send shutdown to all goroutines
			cancel()

			// step: wait for all goroutines to stop
			wg.Wait()
			log.Println("run: all goroutines have stopped, terminating.")
			break listenerloop
			// return
		case err := <-retErrChUnseal:
			log.Printf("run: error received from the vaultUnseal worker: %v", err)
		case err := <-retErrChInit:
			log.Printf("run: error received from the vaultInit worker: %v", err)
			// ECS channels
		}
	}

}

func runEcsDsc(srvconfig DbgConfig, vgconf vaultg.Config) map[string][]string {

	log.Println("ecsw: running ECS discovery")

	// step: discover vault servers: extract type:ECS vault endpoints
	var ecscl []ecs.AwsEcsInput
	for ve := range vgconf.Endpoints {
		for ves := range vgconf.Endpoints[ve].Specs {
			var ae ecs.AwsEcsInput
			ae.Region = vgconf.Endpoints[ve].Specs[ves].Region
			ae.Cluster = vgconf.Endpoints[ve].Specs[ves].Cluster
			ecscl = append(ecscl, ae)
		}
	}
	if debugListenerPtr {
		// for ve := range vgconf.Endpoints {
		log.Printf("config ECS clusters: %v", ecscl)
		// }
		if debugListenerConf {
			spew.Dump(ecscl)
		}
	}

	// step: discover vault servers: pass all ECS endpoints to the ecs pkg for processing
	ecs.PropagateDebug(debugListenerPtr, debugListenerConf)
	dsc := ecs.Discover(ecscl)

	// step: log partial failures
	rdv := make(map[string][]string)
	var dvs []string
	for i := range dsc {

		if len(dsc[i].Fault) != 0 {
			for j := range dsc[i].Fault {
				errm := fmt.Sprintf("listener: cluster discovery error (%v) for cluster: %v", dsc[i].Fault[j], dsc[i].Cluster)
				log.Printf(errm)
			}
			// step: return successful discoveries
		} else {
			for j := range dsc[i].VaultServers {
				ts := fmt.Sprintf("https://%v:%v", dsc[i].VaultServers[j].IP, dsc[i].VaultServers[j].Port)
				dvs = append(dvs, ts)
			}
			rdv[dsc[i].Cluster] = dvs
		}
	}

	return rdv

}

// runHTTPSrv starts the HTTP server
func runHTTPSrv(ctx context.Context, srvConfig DbgConfig, vaultg vaultg.Config, wg *sync.WaitGroup, id workerID) {

	defer wg.Done()
	defer log.Printf("%v%v: gracefully stopped.", id.Name, id.ID)

	addr := vaultg.Address + ":" + vaultg.Port
	logger := log.New(os.Stdout, "", log.Ldate|log.Lshortfile)

	hs := &http.Server{
		Addr:    addr,
		Handler: server.New(server.Logger(logger)),
	}

	go func() {
		logger.Printf("%v%v: server is listening on %v", id.Name, id.ID, hs.Addr)

		if err := hs.ListenAndServe(); err != nil {
			logger.Printf("%v%v: received an error: %v", id.Name, id.ID, err)
		}
	}()

	// long running
	for {
		select {
		case <-ctx.Done():
			if debugListenerPtr {
				log.Printf("%v%v: caller has asked us to stop processing work; gracefully shutting down.", id.Name, id.ID)
			}
			// shut down gracefully, but wait no longer than 5 seconds before halting
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// ignore error since it will be "Err shutting down server : context canceled"
			hs.Shutdown(shutdownCtx)

			return
		}
	}
}
