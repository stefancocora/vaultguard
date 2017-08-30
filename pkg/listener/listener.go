package listener

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/stefancocora/vaultguard/pkg/server"
	vaultg "github.com/stefancocora/vaultguard/pkg/vault"
)

var debugListenerPtr bool

// Config is the type that is used to pass configuration to the http server
type Config struct {
	Address     string `yaml:"listen_address" json:"listen_address"`
	Port        string `yaml:"listen_port" json:"listen_port"`
	Debug       bool
	DebugConfig bool
}

// Entrypoint represents the entrypoint in the server package
func Entrypoint(srvConfig Config) error {
	debugListenerPtr = srvConfig.Debug

	if debugListenerPtr {
		log.Printf("received config: %#v", srvConfig)
	}

	run(srvConfig)

	return nil
}

// run starts all long running threads and communication channels
func run(srvConfig Config) {

	// create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}

	// step: setup OS signal handler
	if debugListenerPtr {
		log.Println("*** debug: run: listening for OS signals on a channel")
	}
	osStopCh := make(chan os.Signal, 1)
	signal.Notify(osStopCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	// step: start the HTTP server
	if debugListenerPtr {
		log.Println("*** debug: run: starting the httpSrv ")
	}
	wg.Add(1)
	go startListen(ctx, srvConfig, wg)

	// step: start vaultInit worker
	if debugListenerPtr {
		vaultg.PropagateDebug(srvConfig.Debug, srvConfig.DebugConfig)
	}
	var vgconf vaultg.Config
	if err := vgconf.New(); err != nil {
		log.Printf("unable to create vaultguard configuration %v", err)
	}
	retErrChInit := make(chan error)
	if vgconf.GuardConfig.Init {
		log.Println("*** debug: run: starting the vaultInit worker")
		wg.Add(1)
		go vaultg.RunInit(ctx, vgconf, wg, retErrChInit) // start vault Init worker
	} else {
		log.Printf("run: init phase is disabled in the config file: %v", vgconf.GuardConfig.Init)
	}

	// step: start vaultUnseal worker
	if debugListenerPtr {
		vaultg.PropagateDebug(srvConfig.Debug, srvConfig.DebugConfig)
	}
	var vgconf02 vaultg.Config
	if err := vgconf02.New(); err != nil {
		log.Printf("unable to create vaultguard configuration %v", err)
	}
	retErrChUnseal := make(chan error)
	if vgconf02.GuardConfig.Init {
		log.Println("*** debug: run: starting the vaultUnseal worker")
		wg.Add(1)
		go vaultg.RunUnseal(ctx, vgconf02, wg, retErrChUnseal) // start vault Init worker
	} else {
		log.Printf("run: unseal phase is disabled in the config file: %v", vgconf.GuardConfig.Init)
	}

	// step: long running wait
	select {
	case sig := <-osStopCh:
		log.Printf("run: received termination signal ( %v ), asking all goroutines to stop.", sig)
		// send shutdown to all goroutines
		cancel()

		// step: wait for all goroutines to stop
		wg.Wait()
		log.Println("run: all goroutines have stopped, terminating.")
	case err := <-retErrChUnseal:
		log.Printf("run: error received from the vaultUnseal worker: %v", err)
	case err := <-retErrChInit:
		log.Printf("run: error received from the vaultInit worker: %v", err)
	}

	return
}

// startListen starts the HTTP server
func startListen(ctx context.Context, srvConfig Config, wg *sync.WaitGroup) {

	defer wg.Done()
	addr := srvConfig.Address + ":" + srvConfig.Port
	logger := log.New(os.Stdout, "", log.Ldate|log.Lshortfile)

	hs := &http.Server{
		Addr:    addr,
		Handler: server.New(server.Logger(logger)),
	}

	go func() {
		logger.Printf("httpSrv: server is listening on %v", hs.Addr)

		if err := hs.ListenAndServe(); err != nil {
			logger.Printf("httpSrv: received an error: %v", err)
		}
	}()

	// long running
	for {
		select {
		case <-ctx.Done():
			if debugListenerPtr {
				log.Println("httpSrv: caller has asked us to stop processing work; gracefully shutting down.")
			}
			// shut down gracefully, but wait no longer than 5 seconds before halting
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// ignore error since it will be "Err shutting down server : context canceled"
			hs.Shutdown(shutdownCtx)

			if debugListenerPtr {
				log.Println("httpSrv: gracefully stopped.")
			}
			return
		}
	}
}
