# Variables
CMD_DIR = ./cmd
MAIN_FILE = $(CMD_DIR)/main.go
BIN = ./bin
BINARY_NAME = image-generator

build:
	@echo "Building the application..."
	go build -o $(BIN)/$(BINARY_NAME) $(MAIN_FILE)
test:
	@echo "Running tests..."
	go test ./...

# Lint command
lint: install-lint
	@echo "Running golangci-lint..."
	golangci-lint run

# Install golangci-lint if not exists
install-lint:
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run command
run:
	@echo "Running the application..."
	go run $(MAIN_FILE)

# TODO add git hooks( pre-commit, pre-push) for linting and testing and go mod tidy etc.