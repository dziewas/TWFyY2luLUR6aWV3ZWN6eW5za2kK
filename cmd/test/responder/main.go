package main

import (
	"crypto/rand"
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

func toAlphaNumLower(b byte) byte {
	const alphabetSize = 'z' - 'a'
	return 'a' + b%alphabetSize
}

func main() {
	var port string

	flag.StringVar(&port, "port", "8080", "http port to listen on")
	flag.Parse()

	r := mux.NewRouter()

	randomSource := rand.Reader
	converter := NewConverter(toAlphaNumLower)
	responder := NewResponder(randomSource, converter)

	r.Handle(rangePath, responder)

	addr := net.JoinHostPort("", port)
	log.Printf("responder listening on %s\n", addr)
	log.Fatalf("server shutdown: %s", http.ListenAndServe(addr, r))
}
