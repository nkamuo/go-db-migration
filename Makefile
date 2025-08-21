# Makefile for Database Migrator

# Variables
BINARY_NAME=migrator
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DIST_DIR=dist

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

# Target platforms
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: all build clean test deps lint help build-all build-windows package-windows package-all

all: clean deps build

## help: Show this help message
help:
	@echo "Database Migrator Build System"
	@echo "Available targets:"
	@echo "  build         Build the binary for current platform"
	@echo "  build-all     Build binaries for all supported platforms"
	@echo "  build-windows Build binaries for Windows platforms only"
	@echo "  package-windows Package Windows binaries into zip files"
	@echo "  package-all   Build and package for all platforms"
	@echo "  test          Run tests"
	@echo "  clean         Clean build artifacts"
	@echo "  deps          Install dependencies"
	@echo "  lint          Run linters"
	@echo "  install       Install binary to /usr/local/bin/"
	@echo ""
	@echo "Supported platforms: linux/amd64, linux/386, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64"

## build: Build the binary for current platform
build:
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/migrator

## build-all: Build for multiple platforms
build-all: clean deps
	@echo "Building for multiple platforms..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/migrator
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/migrator
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/migrator
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/migrator
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/migrator
	GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-arm64.exe ./cmd/migrator
	@echo "All platform builds complete!"

## build-windows: Build for Windows (amd64 and arm64)
build-windows:
	@echo "Building for Windows platforms..."
	@mkdir -p bin
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/migrator
	GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-arm64.exe ./cmd/migrator
	@echo "Windows builds complete!"

## package: Create distribution packages for all platforms
package: build-all
	@echo "Creating distribution packages..."
	@rm -rf $(DIST_DIR)
	@mkdir -p $(DIST_DIR)
	
	# Package for Linux AMD64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64
	@cp bin/$(BINARY_NAME)-linux-amd64 $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/$(BINARY_NAME)
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/conf.json.example
	@echo "#!/bin/bash\n./$(BINARY_NAME) \$$@" > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/run.sh
	@chmod +x $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/run.sh
	cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-$(VERSION)-linux-amd64
	
	# Package for Linux ARM64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64
	@cp bin/$(BINARY_NAME)-linux-arm64 $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/$(BINARY_NAME)
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/conf.json.example
	@echo "#!/bin/bash\n./$(BINARY_NAME) \$$@" > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/run.sh
	@chmod +x $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/run.sh
	cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-$(VERSION)-linux-arm64
	
	# Package for macOS AMD64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64
	@cp bin/$(BINARY_NAME)-darwin-amd64 $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/$(BINARY_NAME)
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/conf.json.example
	@echo "#!/bin/bash\n./$(BINARY_NAME) \$$@" > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/run.sh
	@chmod +x $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/run.sh
	cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-$(VERSION)-darwin-amd64
	
	# Package for macOS ARM64 (Apple Silicon)
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64
	@cp bin/$(BINARY_NAME)-darwin-arm64 $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/$(BINARY_NAME)
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/conf.json.example
	@echo "#!/bin/bash\n./$(BINARY_NAME) \$$@" > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/run.sh
	@chmod +x $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/run.sh
	cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-$(VERSION)-darwin-arm64
	
	# Package for Windows AMD64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64
	@cp bin/$(BINARY_NAME)-windows-amd64.exe $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/$(BINARY_NAME).exe
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/conf.json.example
	@echo "@echo off\n$(BINARY_NAME).exe %*" > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/run.bat
	cd $(DIST_DIR) && zip -r $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-$(VERSION)-windows-amd64
	
	# Package for Windows ARM64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64
	@cp bin/$(BINARY_NAME)-windows-arm64.exe $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/$(BINARY_NAME).exe
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/conf.json.example
	@echo "@echo off\n$(BINARY_NAME).exe %*" > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/run.bat
	cd $(DIST_DIR) && zip -r $(BINARY_NAME)-$(VERSION)-windows-arm64.zip $(BINARY_NAME)-$(VERSION)-windows-arm64
	
	@echo "\nDistribution packages created in $(DIST_DIR)/:"
	@ls -la $(DIST_DIR)/*.tar.gz $(DIST_DIR)/*.zip 2>/dev/null || true

## package-windows: Create Windows distribution packages only
package-windows: build-windows
	@echo "Creating Windows distribution packages..."
	@rm -rf $(DIST_DIR)
	@mkdir -p $(DIST_DIR)
	
	# Package for Windows AMD64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64
	@cp bin/$(BINARY_NAME)-windows-amd64.exe $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/$(BINARY_NAME).exe
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/conf.json.example
	@echo "@echo off" > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/run.bat
	@echo "$(BINARY_NAME).exe %*" >> $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/run.bat
	cd $(DIST_DIR) && zip -r $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-$(VERSION)-windows-amd64
	
	# Package for Windows ARM64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64
	@cp bin/$(BINARY_NAME)-windows-arm64.exe $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/$(BINARY_NAME).exe
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/conf.json.example
	@echo "@echo off" > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/run.bat
	@echo "$(BINARY_NAME).exe %*" >> $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/run.bat
	cd $(DIST_DIR) && zip -r $(BINARY_NAME)-$(VERSION)-windows-arm64.zip $(BINARY_NAME)-$(VERSION)-windows-arm64
	
	@echo "\nWindows distribution packages created in $(DIST_DIR)/:"
	@ls -la $(DIST_DIR)/*.zip 2>/dev/null || true
	
	# Clean up intermediate files and folders
	@echo "Cleaning up intermediate files..."
	@rm -rf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/
	@rm -rf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/
	@rm -f bin/$(BINARY_NAME)-windows-*.exe
	@echo "Cleanup complete!"

## package-all: Build and package for all platforms
package-all: build-all
	@echo "Creating distribution packages for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Package Linux AMD64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64
	@cp bin/$(BINARY_NAME)-linux-amd64 $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/$(BINARY_NAME)
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/conf.json.example
	@echo '#!/bin/bash' > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/run.sh
	@echo './$(BINARY_NAME) "$$@"' >> $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/run.sh
	@chmod +x $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/run.sh
	@cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-$(VERSION)-linux-amd64
	
	# Package Linux 386
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-386
	@cp bin/$(BINARY_NAME)-linux-386 $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-386/$(BINARY_NAME)
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-386/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-386/conf.json.example
	@echo '#!/bin/bash' > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-386/run.sh
	@echo './$(BINARY_NAME) "$$@"' >> $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-386/run.sh
	@chmod +x $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-386/run.sh
	@cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-linux-386.tar.gz $(BINARY_NAME)-$(VERSION)-linux-386
	
	# Package Linux ARM64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64
	@cp bin/$(BINARY_NAME)-linux-arm64 $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/$(BINARY_NAME)
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/conf.json.example
	@echo '#!/bin/bash' > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/run.sh
	@echo './$(BINARY_NAME) "$$@"' >> $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/run.sh
	@chmod +x $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/run.sh
	@cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-$(VERSION)-linux-arm64
	
	# Package macOS AMD64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64
	@cp bin/$(BINARY_NAME)-darwin-amd64 $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/$(BINARY_NAME)
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/conf.json.example
	@echo '#!/bin/bash' > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/run.sh
	@echo './$(BINARY_NAME) "$$@"' >> $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/run.sh
	@chmod +x $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/run.sh
	@cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-$(VERSION)-darwin-amd64
	
	# Package macOS ARM64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64
	@cp bin/$(BINARY_NAME)-darwin-arm64 $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/$(BINARY_NAME)
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/conf.json.example
	@echo '#!/bin/bash' > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/run.sh
	@echo './$(BINARY_NAME) "$$@"' >> $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/run.sh
	@chmod +x $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/run.sh
	@cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-$(VERSION)-darwin-arm64
	
	# Package Windows AMD64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64
	@cp bin/$(BINARY_NAME)-windows-amd64.exe $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/$(BINARY_NAME).exe
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/conf.json.example
	@echo '@echo off' > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/run.bat
	@echo '$(BINARY_NAME).exe %*' >> $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/run.bat
	@cd $(DIST_DIR) && zip -r $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-$(VERSION)-windows-amd64
	
	# Package Windows ARM64
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64
	@cp bin/$(BINARY_NAME)-windows-arm64.exe $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/$(BINARY_NAME).exe
	@cp README.md $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/
	@cp conf.json $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/conf.json.example
	@echo '@echo off' > $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/run.bat
	@echo '$(BINARY_NAME).exe %*' >> $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64/run.bat
	@cd $(DIST_DIR) && zip -r $(BINARY_NAME)-$(VERSION)-windows-arm64.zip $(BINARY_NAME)-$(VERSION)-windows-arm64
	
	@echo "\nAll distribution packages created in $(DIST_DIR)/:"
	@ls -la $(DIST_DIR)/*.{tar.gz,zip} 2>/dev/null || true
	
	# Clean up intermediate files and folders
	@echo "Cleaning up intermediate files..."
	@rm -rf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-*/
	@rm -f bin/$(BINARY_NAME)-*
	@echo "Cleanup complete! All platforms packaged successfully."

## clean: Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf bin/ $(DIST_DIR)/

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
	@echo "Development build complete. Binary in bin/"

## checksums: Generate checksums for distribution files
checksums:
	@if [ -d "$(DIST_DIR)" ]; then \
		echo "Generating checksums..."; \
		cd $(DIST_DIR) && find . -name "*.tar.gz" -o -name "*.zip" | xargs shasum -a 256 > checksums.txt; \
		echo "Checksums saved to $(DIST_DIR)/checksums.txt"; \
	else \
		echo "No distribution directory found. Run 'make package' first."; \
	fi

## release: Build and package everything for release
release: package checksums
	@echo "\nRelease packages ready in $(DIST_DIR)/:"
	@ls -la $(DIST_DIR)/ | grep -E '\.(tar\.gz|zip|txt)$$'
