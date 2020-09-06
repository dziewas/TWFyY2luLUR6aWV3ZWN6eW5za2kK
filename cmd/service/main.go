package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"crawler/pkg/handler"
	"crawler/pkg/store/memory"
)

const (
	defaultLimit = 1024 * 1024
)

func main() {
	var ip, port string
	var limit int

	flag.StringVar(&port, "port", "8080", "http port to listen on")
	flag.StringVar(&ip, "ip", "0.0.0.0", "ip address to bind to")
	flag.IntVar(&limit, "limit", defaultLimit, "payload limit")
	flag.Parse()

	storage := memory.NewMemory()

	fetcher := handler.NewFetcher(storage)
	fetcherStop := fetcher.Start()
	defer fetcherStop()

	router := mux.NewRouter()
	router.Handle("/api/fetcher", http.HandlerFunc(fetcher.Create)).Methods("POST")
	router.Handle("/api/fetcher", http.HandlerFunc(fetcher.List)).Methods("GET")
	router.Handle("/api/fetcher/{id}", http.HandlerFunc(fetcher.Delete)).Methods("DELETE")
	router.Handle("/api/fetcher/{id}/history", http.HandlerFunc(fetcher.History)).Methods("GET")

	sizeLimiter := handler.NewSizeLimiter(limit)

	addr := net.JoinHostPort(ip, port)
	log.Printf("listening on: %s\n", addr)

	err := http.ListenAndServe(addr, handler.NewChain(router, sizeLimiter))
	if err != nil {
		log.Fatalf("starting server failed: %s\n", err)
	}
}
