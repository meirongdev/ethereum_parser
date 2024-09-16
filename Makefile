# Variables
CMD_DIR = ./cmd
MAIN_FILE = $(CMD_DIR)/main.go
BIN = ./bin
BINARY_NAME = image-generator
MOCK_DIRS := \
    ./internal/ethereum

build:
	@echo "Building the application..."
	go build -o $(BIN)/$(BINARY_NAME) $(MAIN_FILE)
test: gen/mocks
	@echo "Running tests..."
	go test -v -race ./...

# Lint command
lint: install-lint
	@echo "Running golangci-lint..."
	golangci-lint run

# Install golangci-lint if not exists
install-lint:
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Mockery command
gen/mocks: install-mockery
	@echo "Running mockery..."
	@rm -rf $(foreach dir, $(MOCK_DIRS), $(dir)/mocks)
	@mockery
install-mockery:
	@echo "Installing mockery..."
	go install github.com/vektra/mockery/v2@latest

# Run command
run:
	@echo "Running the application..."
	go run $(MAIN_FILE)

# TODO add git hooks( pre-commit, pre-push) for linting and testing and go mod tidy etc.