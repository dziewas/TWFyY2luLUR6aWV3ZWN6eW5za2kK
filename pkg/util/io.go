package util

import (
	"io"
	"log"
)

func MustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Printf("failed to close the stream: %s", err)
	}
}
