package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"

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
		log.Fatalf("%s variable not set\n", portEnvVar)
	}

	var limit int

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

	fetcher := handler.NewFetcher(storage, util.GenID)
	fetcherStop := fetcher.Start()
	defer fetcherStop()

	router := handler.NewRouter(fetcher)

	sizeLimiter := handler.NewSizeLimiter(limit)

	addr := net.JoinHostPort("", port)
	log.Printf("listening on: %s\n", addr)

	err := http.ListenAndServe(addr, handler.NewChain(router, sizeLimiter))
	if err != nil {
		log.Fatalf("starting server failed: %s\n", err)
	}
}
