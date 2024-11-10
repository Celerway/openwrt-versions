# Binary name
BINARY=operwrt-version

# Build directory
BUILD_DIR=bin

# Go files
GO_FILES=$(shell find . -name "*.go")

# Linter
LINTER=golangci-lint

# Build the binary
build: $(BUILD_DIR)/$(BINARY)

$(BUILD_DIR)/$(BINARY): $(GO_FILES)
	CGO_ENABLED=0 go build -o $@

# Run the linter
lint:
	$(LINTER) run

# Clean the build directory
clean:
	rm -rf $(BUILD_DIR)

# Run tests
test:
	go test ./...

# Run tests with coverage
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Help target
help:
	@echo "Targets:"
	@echo "  build        Build the binary"
	@echo "  lint         Run the linter"
	@echo "  clean        Clean the build directory"
	@echo "  test         Run tests"
	@echo "  test-cover  Run tests with coverage"
	@echo "  help         Display this help message"

.PHONY: build lint clean test test-cover help