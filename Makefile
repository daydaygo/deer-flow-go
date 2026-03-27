.PHONY: build run test test-api clean fmt lint vet

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -race ./... -v

test-api:
	@./test-api.sh

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

clean:
	rm -rf bin/
	rm -rf .deer-flow/