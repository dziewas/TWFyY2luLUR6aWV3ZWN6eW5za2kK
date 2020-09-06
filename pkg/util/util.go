package util

import (
	"io"
	"log"
	"math/rand"
	"time"
)

func inti() {
	rand.Seed(time.Now().UnixNano())
}

func NewID() int {
	return rand.Int()
}

func MustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Printf("failed to close the stream: %s", err)
	}
}

func NowFunc() time.Time {
	return time.Now()
}
