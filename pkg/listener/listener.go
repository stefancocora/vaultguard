package listener

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/stefancocora/vaultguard/pkg/server"
	"github.com/stefancocora/vaultguard/pkg/vault"
)

var debugListenerPtr bool

// Config is the type that is used to pass configuration to the http server
type Config struct {
	Address string `yaml:"listen_address" json:"listen_address"`
	Port    string `yaml:"listen_port" json:"listen_port"`
	Debug   bool
}

// Entrypoint represents the entrypoint in the server package
func Entrypoint(srvConfig Config) error {
	debugListenerPtr = srvConfig.Debug

	if debugListenerPtr {
		log.Printf("received config: %#v", srvConfig)
	}

	if err := startListen(srvConfig); err != nil {
		return errors.Wrap(err, "server unable to start listening: %v")
	}

	if err := vault.Entrypoint(); err != nil {
		return errors.Wrap(err, "vault handler error")
	}

	return nil
}

// startListen starts the HTTP server
func startListen(srvConfig Config) error {

	hs, logger := setup(srvConfig)

	go func() {
		logger.Printf("server is listening on %v", hs.Addr)
		if err := hs.ListenAndServe(); err != nil {
			logger.Fatalf("unable to bind to %v", hs.Addr)
		}
	}()

	graceful(hs, logger, 5*time.Second)

	return nil
}

func setup(srvConfig Config) (*http.Server, *log.Logger) {
	addr := srvConfig.Address + ":" + srvConfig.Port

	logger := log.New(os.Stdout, "", 0)

	return &http.Server{
		Addr:    addr,
		Handler: server.New(server.Logger(logger)),
	}, logger
}

func graceful(hs *http.Server, logger *log.Logger, timeout time.Duration) {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Printf("\nShutdown with timeout: %s\n", timeout)

	if err := hs.Shutdown(ctx); err != nil {
		logger.Printf("Error: %v\n", err)
	} else {
		logger.Println("Server stopped")
	}
}
