package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Responder struct{}

func NewResponder() *Responder {
	return &Responder{}
}

func (r *Responder) sendChunk(sink http.ResponseWriter, converter io.ReadWriter, source io.Reader, chunk int64) error {
	_, err := io.CopyN(converter, source, chunk)
	if err != nil {
		return err
	}

	_, err = io.CopyN(sink, converter, chunk)
	if err != nil {
		return err
	}

	return nil
}

func (r *Responder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sizeRaw, ok := mux.Vars(req)[sizeParam]
	if !ok {
		http.Error(w, "size param not specified", http.StatusBadRequest)
		return
	}

	size, err := strconv.ParseInt(sizeRaw, 10, 64)
	if err != nil {
		http.Error(w, "size param invalid", http.StatusBadRequest)
		return
	}

	if size < 0 || size > sizeLimit {
		http.Error(w, fmt.Sprintf("size param not in range (%d, %d)", 0, sizeLimit), http.StatusBadRequest)
		return
	}

	chunk := int64(defaultChunk)
	if size < defaultChunk {
		chunk = size
	}

	randomSource := rand.Reader
	converter := NewConverter(toAlphaNumLower)

	for i := int64(0); i < size/chunk; i++ {
		err := r.sendChunk(w, converter, randomSource, chunk)
		if err != nil {
			log.Printf("sending chunk failed: %s\n", err)
		}
	}

	chunk = size % chunk
	if chunk > 0 {
		err := r.sendChunk(w, converter, randomSource, chunk)
		if err != nil {
			log.Printf("sending chunk failed: %s\n", err)
		}
	}
}

func toAlphaNumLower(b byte) byte {
	const alphabetSize = 'z' - 'a'
	return 'a' + b%alphabetSize
}
