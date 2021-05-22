TEST=go test -v -cover -race -count=1 -timeout 30s $(1)

.SILENT:
.PHONY: build
build:
	go build -v ./cmd/apiserver

build-static:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v -a ./cmd/apiserver

.PHONY: test
test:
	$(call TEST,./internal/app/apiserver/...)
	$(call TEST,./internal/app/store/...)

.DEFAULT_GOAL := build