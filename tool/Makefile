.PHONY: build install test clean validate fmt lint help

BINARY := minispec
BUILD_DIR := bin
CMD_DIR := cmd/minispec
VERSION := $(shell grep '^Version:' ../.claude/skills/mini-spec/README.md | cut -d' ' -f2)
LDFLAGS := -ldflags "-X github.com/zot/minispec/internal/cli.Version=$(VERSION)"

# Build the binary
build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./$(CMD_DIR)

# Install to ~/.claude/bin
install: build
	mkdir -p ~/.claude/bin
	cp $(BUILD_DIR)/$(BINARY) ~/.claude/bin/

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run minispec validate on this project
validate: build
	./$(BUILD_DIR)/$(BINARY) validate

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy

# Build for multiple platforms
release:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./$(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./$(CMD_DIR)

help:
	@echo "minispec build targets:"
	@echo "  build          Build the binary to bin/"
	@echo "  install        Install to ~/.claude/bin"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  clean          Remove build artifacts"
	@echo "  validate       Run minispec validate on this project"
	@echo "  fmt            Format code"
	@echo "  lint           Run linter (requires golangci-lint)"
	@echo "  tidy           Tidy go.mod dependencies"
	@echo "  release        Build for multiple platforms"
	@echo "  help           Show this help"
