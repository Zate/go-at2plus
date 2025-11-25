BINARY_NAME=at2plus
GO_FILES=$(shell find . -name '*.go')

.PHONY: all build test lint clean

all: build test lint

build:
	go build -o bin/$(BINARY_NAME) ./cmd/at2plus

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf bin
	go clean
