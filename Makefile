.PHONY: build

.SILENT:
build:
	go build -v ./cmd/apiserver

.PHONY: test
test:
	go test -v -race -count=1 -timeout 30s ./internal/...

.DEFAULT_GOAL := build