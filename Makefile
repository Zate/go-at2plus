BINARY_NAME=at2plus
GO_FILES=$(shell find . -name '*.go')

.PHONY: all fmt vet lint test build clean

all: fmt vet lint test build

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

test:
	go test -v ./...

build:
	go build -o bin/$(BINARY_NAME) ./cmd/at2plus

clean:
	rm -rf bin
	go clean
