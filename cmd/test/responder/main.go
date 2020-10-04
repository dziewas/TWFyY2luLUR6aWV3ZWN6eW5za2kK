package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

const (
	defaultChunk = 256
	sizeLimit    = 1024 * 1024
	sizeParam    = "size"
	rangePath    = "/range/{" + sizeParam + "}"
	portEnvVar   = "PORT"
)

func main() {
	port := os.Getenv(portEnvVar)
	if port == "" {
		log.Fatalf("%s variable not set\n", portEnvVar)
	}

	r := mux.NewRouter()

	responder := NewResponder()

	r.Handle(rangePath, responder)

	addr := net.JoinHostPort("", port)
	log.Printf("responder listening on %s\n", addr)
	log.Fatalf("server shutdown: %s", http.ListenAndServe(addr, r))
}
