package handler

import (
	"log"
	"net/http"
	"strconv"
)

func NewChain(f http.Handler, handlers ...func(http.Handler) http.Handler) http.Handler {
	for i := len(handlers) - 1; i >= 0; i-- {
		handler := handlers[i]

		f = handler(f)
	}

	return f
}

func NewSizeLimiter(limit int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					contentLength, err := strconv.Atoi(r.Header.Get("Content-Length"))
					if err != nil {
						log.Println("checking incoming content size failed")
						http.Error(w, "internal error", http.StatusInternalServerError)
						return
					}

					if contentLength > limit {
						http.Error(w, "", http.StatusRequestEntityTooLarge)
						return
					}
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}
