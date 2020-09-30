package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	defaultChunk = 256
	sizeLimit    = 1024 * 1024
	sizeParam    = "size"
	rangePath    = "/range/{" + sizeParam + "}"
)

func main() {
	var port string

	flag.StringVar(&port, "port", "8080", "http port to listen on")
	flag.Parse()

	r := mux.NewRouter()

	responder := NewResponder()

	r.Handle(rangePath, responder)

	addr := net.JoinHostPort("", port)
	log.Printf("responder listening on %s\n", addr)
	log.Fatalf("server shutdown: %s", http.ListenAndServe(addr, r))
}
