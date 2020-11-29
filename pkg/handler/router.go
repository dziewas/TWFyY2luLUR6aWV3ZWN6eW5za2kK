package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func dummy(_ http.ResponseWriter, _ *http.Request) {
}

func NewRouter(fetcher *Fetcher) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/api/fetcher", http.HandlerFunc(fetcher.Create)).Methods("POST")
	router.Handle("/api/fetcher", http.HandlerFunc(fetcher.List)).Methods("GET")
	router.Handle("/api/fetcher/{id}", http.HandlerFunc(fetcher.Delete)).Methods("DELETE")
	router.Handle("/api/fetcher/{id}/history", http.HandlerFunc(fetcher.History)).Methods("GET")

	router.HandleFunc("/api/fetcher", dummy).Methods("OPTIONS")
	router.HandleFunc("/api/fetcher/{id}", dummy).Methods("OPTIONS")
	router.HandleFunc("/api/fetcher/{id}/history", dummy).Methods("OPTIONS")

	return router
}
