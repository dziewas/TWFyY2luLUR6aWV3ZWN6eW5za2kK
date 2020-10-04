package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(fetcher *Fetcher) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/api/fetcher", http.HandlerFunc(fetcher.Create)).Methods("POST")
	router.Handle("/api/fetcher", http.HandlerFunc(fetcher.List)).Methods("GET")
	router.Handle("/api/fetcher/{id}", http.HandlerFunc(fetcher.Delete)).Methods("DELETE")
	router.Handle("/api/fetcher/{id}/history", http.HandlerFunc(fetcher.History)).Methods("GET")

	return router
}
