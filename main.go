package main

import (
	"log"

	"github.com/pkg/errors"
	"github.com/stefancocora/vaultguard/cmd"
	"github.com/stefancocora/vaultguard/pkg/version"
)

var binversion string

func main() {
	vers, err := version.Printvers()
	if err != nil {
		log.Fatalf("[FATAL] main - %v", errors.Cause(err))
	}

	binversion = vers
	log.Printf("bin: %v binver: %v pkg: %v", version.BinaryName, vers, "main")

	log.Print("starting engines")
	buildctx, err := version.BuildContext()
	if err != nil {
		log.Fatalf("[FATAL] main wrong BuildContext - %v", errors.Cause(err))
	}

	log.Printf("build context ( %v )", buildctx)

	cmd.Execute()

	log.Print("stopping engines, we're done")
}

// func main() {

// 	// handlers
// 	http.HandleFunc("/healthz", healthz)       // new goroutine, careful with locking
// 	http.HandleFunc("/status", status)         // new goroutine, careful with locking
// 	http.HandleFunc("/pausewatch", pausewatch) // new goroutine, careful with locking

// 	addr := "localhost:8001"
// 	log.Printf("%v is listening on %v\n", binary, addr)
// 	if err := http.ListenAndServe(addr, nil); err != nil {
// 		log.Fatalf("[FATAL] unable to bind to %v: %v", addr, err)
// 	} // attaching to DefaultServeMux

// }

// func healthz(res http.ResponseWriter, req *http.Request) {

// 	switch req.Method {
// 	case "GET":
// 		res.WriteHeader(http.StatusOK)
// 		fmt.Fprint(res, "health: ok")
// 	default:
// 		http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
// 	}

// }

// // status should heartbeat and check multiple vault servers for
// // if initialized
// // if unsealed
// func status(res http.ResponseWriter, req *http.Request) {

// 	switch req.Method {
// 	case "GET":
// 		res.WriteHeader(http.StatusOK)
// 		fmt.Fprint(res, "status: doing nothing for now")
// 	default:
// 		http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
// 	}
// }

// // pausewatch should put the webserver into a wait like state where it will respond with /healthz 200Ok and body { pause activated, not watching over any vault instances }
// func pausewatch(res http.ResponseWriter, req *http.Request) {

// 	switch req.Method {
// 	case "GET":
// 		res.WriteHeader(http.StatusOK)
// 		fmt.Fprint(res, "pausewatch: doing nothing for now")
// 	default:
// 		http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
// 	}
// }
