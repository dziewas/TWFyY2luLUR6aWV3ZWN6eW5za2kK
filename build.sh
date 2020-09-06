#!/usr/bin/env bash

go test -v ./... &&
  go build -v ./... &&
  go mod tidy && goimports -w -local github.com/dziewas/interview-wp cmd pkg
