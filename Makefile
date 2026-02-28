# Makefile for agenta11y

.PHONY: all build test lint clean install demo-reports demo-w3c-bad demo-accesscomputing demo-a11yquest-forms

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=agenta11y
BINARY_DIR=./cmd/agent-a11y

# Default target
all: lint test build

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) $(BINARY_DIR)

# Install to GOPATH/bin
install:
	$(GOCMD) install $(BINARY_DIR)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Update dependencies
deps:
	$(GOMOD) tidy
	$(GOMOD) download

# Generate all demo site reports
demo-reports: build
	./$(BINARY_NAME) demo generate

# Generate W3C BAD demo reports
demo-w3c-bad: build
	./$(BINARY_NAME) demo generate w3c-bad

# Generate AccessComputing demo reports
demo-accesscomputing: build
	./$(BINARY_NAME) demo generate accesscomputing

# Generate A11yQuest Forms demo reports
demo-a11yquest-forms: build
	./$(BINARY_NAME) demo generate a11yquest

# Run a single audit
# Usage: make audit URL=https://example.com
audit: build
	./$(BINARY_NAME) audit $(URL)

# Run a comparison
# Usage: make compare BEFORE=https://example.com/before AFTER=https://example.com/after
compare: build
	./$(BINARY_NAME) compare $(BEFORE) $(AFTER)

# Help
help:
	@echo "agenta11y Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build, test, and lint"
	@echo "  make build        Build the binary"
	@echo "  make install      Install to GOPATH/bin"
	@echo "  make test         Run tests"
	@echo "  make lint         Run linter"
	@echo "  make clean        Remove build artifacts"
	@echo "  make deps         Update dependencies"
	@echo ""
	@echo "Demo Reports:"
	@echo "  make demo-reports         Generate all demo site reports"
	@echo "  make demo-w3c-bad         Generate W3C BAD reports"
	@echo "  make demo-accesscomputing Generate AccessComputing reports"
	@echo "  make demo-a11yquest-forms Generate A11yQuest Forms reports"
	@echo ""
	@echo "Single Operations:"
	@echo "  make audit URL=<url>                Run audit on a URL"
	@echo "  make compare BEFORE=<url> AFTER=<url>  Compare two URLs"
