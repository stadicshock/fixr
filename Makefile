BINARY := fixr
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build run clean test install help

## build: Build the fixr binary
build:
	GOTOOLCHAIN=local go build $(LDFLAGS) -o $(BINARY) .

## run: Build and run fixr
run: build
	./$(BINARY)

## install: Install fixr to $$GOPATH/bin
install:
	GOTOOLCHAIN=local go install $(LDFLAGS) .

## test: Run tests
test:
	GOTOOLCHAIN=local go test ./... -v

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)
	go clean

## lint: Run linter
lint:
	golangci-lint run ./...

## tidy: Tidy go modules
tidy:
	GOTOOLCHAIN=local go mod tidy

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
