package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"

	"crawler/pkg/store"

	"github.com/gorilla/mux"

	"crawler/pkg/handler"
	"crawler/pkg/store/memory"
	redis_db "crawler/pkg/store/redis"
)

const (
	defaultLimit = 1024 * 1024
	redisEnvVar  = "REDIS_URL"
)

func main() {
	var ip, port string
	var limit int

	flag.StringVar(&port, "port", "8080", "http port to listen on")
	flag.StringVar(&ip, "ip", "0.0.0.0", "ip address to bind to")
	flag.IntVar(&limit, "limit", defaultLimit, "payload limit")
	flag.Parse()

	var storage store.Store

	redisUrl := os.Getenv(redisEnvVar)
	if len(redisUrl) == 0 {
		log.Printf("'%s' env var not set, using in-mem Store\n", redisEnvVar)
		storage = memory.NewMemory()
	} else {
		opts, err := redis.ParseURL(redisUrl)
		if err != nil {
			log.Fatalf("parsing redis url failed: %s\n", err)
		}

		rdb := redis.NewClient(opts)
		defer func() {
			err := rdb.Close()
			if err != nil {
				log.Fatalf("closing redis connection failed: %s", err)
			}
		}()

		log.Println("redis ping...")
		pong, err := rdb.Ping(context.Background()).Result()
		if err != nil {
			log.Fatalf("redis not responsive: %s", err)
		}

		log.Println(pong, err)

		storage = redis_db.NewStore(rdb)
	}

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
