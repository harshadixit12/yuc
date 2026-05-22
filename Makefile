BINARY_NAME=yuc

.PHONY: all build test lint format clean

all: build

build:
	go build -o $(BINARY_NAME) main.go

test:
	go test -v ./...

lint:
	golangci-lint run

format:
	go fmt ./...

clean:
	rm -f $(BINARY_NAME)
	go clean
