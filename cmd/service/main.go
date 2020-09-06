package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"wp/crawler/pkg/handler"
)

func main() {
	var ip, port string

	flag.StringVar(&port, "port", "8080", "http port to listen on")
	flag.StringVar(&ip, "ip", "0.0.0.0", "ip address to bind to")
	flag.Parse()

	var helloHandler handler.Hello

	r := mux.NewRouter()
	r.Handle("/hello", &helloHandler).Methods("POST")

	addr := net.JoinHostPort(ip, port)
	log.Printf("listening on: %s\n", addr)

	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("starting server failed: %s\n", err)
	}
}
