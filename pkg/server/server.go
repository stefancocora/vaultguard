/*
Copyright 2015 All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var debugSrvPtr bool

// Server is the struct representing the state in rAM of a running server
type Server struct {
	logger *log.Logger
	mux    *http.ServeMux
}

// New creates an instance of a mux server
func New(options ...func(*Server)) *Server {
	s := &Server{mux: http.NewServeMux()}

	for _, f := range options {
		f(s)
	}

	if s.logger == nil {
		// s.logger = log.New(os.Stdout, "", 0)
		s.logger = log.New(os.Stdout, "", log.Ldate|log.Lshortfile)
	}

	s.mux.HandleFunc("/healthz", s.healthz)
	s.mux.HandleFunc("/status", s.status)
	s.mux.HandleFunc("/pausewatch", s.pausewatch)

	return s
}

// ServeHTTP is the http handler for the mux
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "go server")

	s.mux.ServeHTTP(w, r)
}

// Logger is the server logger
func Logger(logger *log.Logger) func(*Server) {
	return func(s *Server) {
		s.logger = logger
	}
}

// HTTP handlers

func (s *Server) healthz(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "GET":
		stc := http.StatusOK
		res.WriteHeader(stc)
		fmt.Fprint(res, "health: ok")
		s.logger.Printf("%v %v %v %v %v", req.RemoteAddr, req.Method, req.URL.Path, req.Proto, stc)
	default:
		http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
	}

}

// status should heartbeat and check multiple vault servers for
// if initialized
// if unsealed
func (s *Server) status(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "GET":
		stc := http.StatusOK
		res.WriteHeader(stc)
		fmt.Fprint(res, "status: doing nothing for now")
		s.logger.Printf("%v %v %v %v %v", req.RemoteAddr, req.Method, req.URL.Path, req.Proto, stc)
	default:
		http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
	}
}

// pausewatch should put the webserver into a wait like state where it will respond with /healthz 200Ok and body { pause activated, not watching over any vault instances }
func (s *Server) pausewatch(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "PUT":
		stc := http.StatusOK
		res.WriteHeader(stc)
		fmt.Fprint(res, "paused watching vault servers")
		s.logger.Printf("%v %v %v %v %v", req.RemoteAddr, req.Method, req.URL.Path, req.Proto, stc)
	default:
		http.Error(res, "Only PUT is allowed", http.StatusMethodNotAllowed)
	}
}
