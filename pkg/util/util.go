package util

import (
	rand_strong "crypto/rand"
	"io"
	"log"
	"math/big"
	rand_weak "math/rand"
	"time"
)

func GenID(maxID int64) int64 {
	id, err := rand_strong.Int(rand_strong.Reader, big.NewInt(maxID))
	if err != nil {
		log.Printf("random generator failed, fallback to default generator")

		return int64(rand_weak.Intn(int(maxID)))
	}

	return id.Int64()
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
