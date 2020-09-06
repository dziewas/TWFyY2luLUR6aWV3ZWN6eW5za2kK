package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Responder struct {
	source    io.Reader
	converter io.ReadWriter
}

func NewResponder(source io.Reader, converter io.ReadWriter) *Responder {
	return &Responder{source: source, converter: converter}
}

func (r *Responder) sendChunk(w http.ResponseWriter, chunk int64) error {
	_, err := io.CopyN(r.converter, r.source, chunk)
	if err != nil {
		return err
	}

	_, err = io.CopyN(w, r.converter, chunk)
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

	for i := int64(0); i < size/chunk; i++ {
		err := r.sendChunk(w, chunk)
		if err != nil {
			log.Printf("sending chunk failed: %s\n", err)
		}
	}

	chunk = size % chunk
	if chunk > 0 {
		err := r.sendChunk(w, chunk)
		if err != nil {
			log.Printf("sending chunk failed: %s\n", err)
		}
	}
}
