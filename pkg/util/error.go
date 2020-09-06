package util

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	err error
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e Error) Error() string {
	return e.err.Error()
}

func Wrap(err error, msg string) *Error {
	return &Error{err: fmt.Errorf("%w: %s", err, msg)}
}

func EmitHttpError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrResourceNotFound) {
		http.Error(w, "", http.StatusNotFound)
	} else if errors.Is(err, ErrValidation) {
		http.Error(w, "", http.StatusBadRequest)
	} else {
		http.Error(w, "", http.StatusInternalServerError)
	}
}

var (
	ErrResourceNotFound = errors.New("resource not found")
	ErrValidation       = errors.New("invalid request")
)
