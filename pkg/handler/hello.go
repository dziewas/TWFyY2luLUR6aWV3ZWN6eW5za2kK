package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"wp/crawler/pkg/model"
	"wp/crawler/pkg/util"
)

type Hello struct {
}

func (h *Hello) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	var person model.Person

	err := d.Decode(&person)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

	defer util.MustClose(r.Body)

	_, err = fmt.Fprintf(w, "Hello %s\n", person.Name)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
