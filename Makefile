.PHONY: build

.SILENT:
build:
	go build -v ./cmd/apiserver

build-static:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v -a ./cmd/apiserver

.PHONY: test
test:
	go test -v -cover -race -count=1 -timeout 30s ./internal/app/apiserver/...
	go test -v -cover -race -count=1 -timeout 30s ./internal/app/store/sqlstore/...

.DEFAULT_GOAL := build