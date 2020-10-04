#!/usr/bin/env bash

# tidy up
go test -v ./... &&
  go build -v ./... &&
  go fmt ./... &&
  go mod tidy && goimports -w -local crawler cmd pkg &&
  gofmt -l -e -d -s cmd/ pkg/

# build
mkdir -p bin &&
  go build -o bin/crawler cmd/service/main.go &&
  go build -o bin/responder cmd/test/responder/*.go

# docker
docker-compose build
