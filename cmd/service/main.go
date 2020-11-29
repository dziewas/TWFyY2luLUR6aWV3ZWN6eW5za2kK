package main

import (
	"context"
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"

	"crawler/pkg/handler"
	"crawler/pkg/store"
	"crawler/pkg/store/memory"
	redis_db "crawler/pkg/store/redis"
	"crawler/pkg/util"
)

const (
	defaultLimit = 1024 * 256
	redisEnvVar  = "REDIS_URL"
	portEnvVar   = "PORT"
)

func main() {
	port := os.Getenv(portEnvVar)
	if port == "" {
		log.Fatalf("%s variable not set", portEnvVar)
	}

	var limit int

	flag.IntVar(&limit, "limit", defaultLimit, "payload limit")
	flag.Parse()

	var storage store.Store

	redisUrl := os.Getenv(redisEnvVar)
	if len(redisUrl) == 0 {
		log.Printf("'%s' env var not set, using in-mem Store", redisEnvVar)
		storage = memory.NewMemory()
	} else {
		opts, err := redis.ParseURL(redisUrl)
		if err != nil {
			log.Fatalf("parsing redis url failed: %s", err)
		}

		rdb := redis.NewClient(opts)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		err = util.RedisConnect(ctx, rdb)
		if err != nil {
			log.Fatalf("timeout waiting for redis: %s", err)
		}

		defer util.MustClose(rdb)

		storage = redis_db.NewStore(rdb)
	}

	fetcher := handler.NewFetcher(storage, util.GenID)
	fetcherStop := fetcher.Start()
	defer fetcherStop()

	router := handler.NewRouter(fetcher)
	cors := mux.CORSMethodMiddleware(router)
	sizeLimiter := handler.NewSizeLimiter(limit)
	contentType := handler.NewContentTypeMW()

	addr := net.JoinHostPort("", port)
	log.Printf("listening on: %s", addr)

	err := http.ListenAndServe(addr, handler.NewChain(router, contentType, sizeLimiter, cors))
	if err != nil {
		log.Fatalf("starting server failed: %s", err)
	}
}
