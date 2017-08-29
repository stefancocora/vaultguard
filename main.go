package main

import (
	"fmt"
	"log"
	"net/http"
)

const binary = "vaultguard"

func main() {

	// handlers
	http.HandleFunc("/healthz", healthz)       // new goroutine, careful with locking
	http.HandleFunc("/status", status)         // new goroutine, careful with locking
	http.HandleFunc("/pausewatch", pausewatch) // new goroutine, careful with locking

	addr := "localhost:8001"
	log.Printf("%v is listening on %v\n", binary, addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("[FATAL] unable to bind to %v: %v", addr, err)
	} // attaching to DefaultServeMux

}

func healthz(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "GET":
		res.WriteHeader(http.StatusOK)
		fmt.Fprint(res, "health: ok")
	default:
		http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
	}

}

// status should heartbeat and check multiple vault servers for
// if initialized
// if unsealed
func status(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "GET":
		res.WriteHeader(http.StatusOK)
		fmt.Fprint(res, "status: doing nothing for now")
	default:
		http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
	}
}

// pausewatch should put the webserver into a wait like state where it will respond with /healthz 200Ok and body { pause activated, not watching over any vault instances }
func pausewatch(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "GET":
		res.WriteHeader(http.StatusOK)
		fmt.Fprint(res, "pausewatch: doing nothing for now")
	default:
		http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
	}
}
