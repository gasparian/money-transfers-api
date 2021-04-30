.PHONY: build
.SILENT:

build:
	go build -v ./cmd/apiserver

.DEFAULT_GOAL := build