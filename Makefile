.PHONY: help run build test clean dev

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

run: ## Run the server
	@echo "Starting Go Initializer on http://localhost:8080"
	@go run main.go

build: ## Build the server
	@echo "Building..."
	@go build -o bin/go-initializer main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/

dev: ## Run with hot reload (requires air)
	@air

install-air: ## Install air for hot reload
	@go install github.com/cosmtrek/air@latest

.DEFAULT_GOAL := help
