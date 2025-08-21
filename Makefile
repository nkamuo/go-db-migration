# Makefile for Database Migrator

# Variables
BINARY_NAME=migrator
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X 'github.com/nkamuo/go-db-migration/internal/cli.Version=$(VERSION)' \
                 -X 'github.com/nkamuo/go-db-migration/internal/cli.BuildDate=$(BUILD_DATE)' \
                 -X 'github.com/nkamuo/go-db-migration/internal/cli.GitCommit=$(GIT_COMMIT)'"

.PHONY: all build clean test deps lint help

all: clean deps build

## build: Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/migrator

## build-all: Build for multiple platforms
build-all: clean deps
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe .

## clean: Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf bin/

## test: Run tests
test:
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## deps: Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## lint: Run linters
lint:
	golangci-lint run

## format: Format code
format:
	gofmt -s -w .
	goimports -w .

## install: Install the binary
install: build
	cp bin/$(BINARY_NAME) /usr/local/bin/

## dev: Build and copy schema for development
dev: build
	cp schema.json bin/
	@echo "Development build complete. Binary and schema in bin/"

## help: Show this help
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## run-example: Run example commands (requires valid database)
run-example: build
	@echo "Testing connection..."
	./bin/$(BINARY_NAME) test-connection
	@echo "\nValidating schema file..."
	./bin/$(BINARY_NAME) validate-schema
	@echo "\nRunning validation checks..."
	./bin/$(BINARY_NAME) validate-all --format table
