package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	// "strconv"
	"strings"
)

// ShiftPath splits off the first component of p, which will be cleaned of
// relative components before processing. head will never contain a slash and
// tail will always be a rooted path without trailing slash.
func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

// A API is doing something
type API struct {
	// We could use http.Handler as a type here; using the specific type has
	// the advantage that static analysis tools can link directly from
	// h.HealthzHandler.ServeHTTP to the correct definition. The disadvantage is
	// that we have slightly stronger coupling. Do the tradeoff yourself.
	// https://blog.merovius.de/2017/06/18/how-not-to-use-an-http-router.html
	HealthzHandler *HealthzHandler
}

func (h *API) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	// var tail string
	// head, tail = ShiftPath(req.URL.String())
	// fmt.Printf("head: %s, tail: %s\n", head, tail)
	head, _ = ShiftPath(req.URL.String())
	// head, req.URL.String() = ShiftPath(req.URL.String())
	if head == "healthz" {
		h.HealthzHandler.ServeHTTP(res, req)
		return
	}
	http.Error(res, "Not Found", http.StatusNotFound)
}

// A HealthzHandler is doing something
type HealthzHandler struct {
}

func (h *HealthzHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "GET":
		h.healthzHandleGet(res, req)
	case "PUT":
		h.handlePut(res, req)
	default:
		http.Error(res, "Only GET and PUT are allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HealthzHandler) healthzHandleGet(res http.ResponseWriter, req *http.Request) {
  res.WriteHeader(http.StatusOK)
	fmt.Fprintln(res, "health: ok")
}

func (h *HealthzHandler) handlePut(res http.ResponseWriter, req *http.Request) {
	fmt.Println("handle PUT method")
}

func main() {
	bindAddr := "0.0.0.0:8001"
	a := &API{
		HealthzHandler: new(HealthzHandler),
	}
	if err := http.ListenAndServe(bindAddr, a); err != nil {
		log.Fatalf("[FATAL] unable to bind to %v: %v", bindAddr, err)
	}
}
