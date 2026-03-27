.PHONY: build run test clean fmt lint vet

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -race ./... -v

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

clean:
	rm -rf bin/
	rm -rf .deer-flow/