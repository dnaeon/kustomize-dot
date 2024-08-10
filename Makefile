.DEFAULT_GOAL := build
LOCAL_BIN ?= $(shell pwd)/bin

BINARY ?= $(LOCAL_BIN)/kustomize-dot

$(LOCAL_BIN):
	mkdir -p $(LOCAL_BIN)

$(BINARY): $(LOCAL_BIN)
	go build -o $(BINARY) ./cmd/kustomize-dot/

build: $(BINARY)

tidy:
	go mod tidy

test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

docker-build:
	docker build -t dnaeon/kustomize-dot:latest .

.PHONY: build tidy test test-cover docker-build
