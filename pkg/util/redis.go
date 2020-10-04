package util

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func RedisConnect(ctx context.Context, rdb *redis.Client) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.New("timeout while waiting for redis")
		case <-ticker.C:
			log.Println("redis ping...")
			pong, err := rdb.Ping(context.Background()).Result()
			if err != nil {
				continue
			}

			log.Println(pong)

			return nil
		}
	}
}
