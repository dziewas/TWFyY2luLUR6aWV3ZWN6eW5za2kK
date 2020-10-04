// +build !unit integration

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"crawler/pkg/util"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	redisUrl           = "redis://redis:6379"
	responderUrlFormat = "http://responder:8080/range/%d"
	createTaskUrl      = "http://crawler:8080/api/fetcher"
	getTaskUrlFormat   = "http://crawler:8080/api/fetcher/%s/history"
)

var (
	rdb *redis.Client
)

func TestMain(m *testing.M) {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		log.Fatalf("parsing redis url failed: %s", err)
	}

	rdb = redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = util.RedisConnect(ctx, rdb)
	if err != nil {
		log.Fatalf("timeout waiting for redis: %s", err)
	}

	defer util.MustClose(rdb)

	os.Exit(m.Run())
}

func TestFull(t *testing.T) {
	const (
		responseSize           = 100
		contentType            = "application/json"
		expectedHistoryEntries = 5
	)

	url := fmt.Sprintf(responderUrlFormat, responseSize)
	payload := fmt.Sprintf(`
		{
			"url": "%s",
			"interval": 1
		}`, url)

	resp, err := http.Post(createTaskUrl, contentType, strings.NewReader(payload))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	id := resp.Header.Get("Location")

	// wait some time until some responses are available
	var history []interface{}
	for i := 0; i < expectedHistoryEntries*2; i++ {
		time.Sleep(time.Second)

		resp, err = http.Get(fmt.Sprintf(getTaskUrlFormat, id))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		err := json.NewDecoder(resp.Body).Decode(&history)
		require.NoError(t, err)

		log.Printf("task %s history has now %d entries", id, len(history))
		if len(history) >= expectedHistoryEntries {
			break
		}
	}

	for _, historyEntryRaw := range history {
		historyEntry, ok := historyEntryRaw.(map[string]interface{})
		require.True(t, ok)

		response, ok := historyEntry["response"].(string)
		require.True(t, ok)

		assert.Equal(t, responseSize, len(response))
	}
}
