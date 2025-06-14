.PHONY: setup test build clean install lint format help

# Default target
help: ## Show this help message
	@echo "DBC GoLang SDK - Available commands:"
	@echo "===================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $1, $2}'

# Setup and installation
setup: ## Set up development environment
	@echo "🚀 Setting up DBC GoLang SDK development environment..."
	@./scripts/setup_testnet.sh

install: ## Install Go dependencies
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy

install-tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/govulncheck@latest

# Development
format: ## Format code
	@echo "🎨 Formatting code..."
	@gofmt -w .
	@goimports -w .

lint: ## Run linter
	@echo "🔍 Running linter..."
	@golangci-lint run ./...

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@go vet ./...

security: ## Run security checks
	@echo "🔒 Running security checks..."
	@govulncheck ./...

# Testing
test: ## Run all tests
	@echo "🧪 Running all tests..."
	@./scripts/run_tests.sh

test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	@go test ./pkg/dbc -v

test-integration: ## Run integration tests only
	@echo "🧪 Running integration tests..."
	@go test ./examples -v -tags=integration

run-meteora-math: ## Run Meteora mathematical functions test
	@echo "🧮 Testing Meteora mathematical functions..."
	@go run examples/meteora/test_math_functions.go

run-meteora-pool: ## Run Meteora pool creation example
	@echo "🏊 Testing Meteora pool creation..."
	@go run examples/meteora/create_pool_and_swap.go

run-meteora-info: ## Run Meteora pool info example
	@echo "📊 Testing Meteora pool information..."
	@go run examples/meteora/get_pool_info.go

run-meteora-all: run-meteora-math run-meteora-pool run-meteora-info ## Run all Meteora examples

# Assessment Examples
run-assessment-config: ## Run CreateConfig assessment example
	@echo "🔧 Running CreateConfig assessment example..."
	@go run examples/assessment/01_create_config.go

run-assessment-pool: ## Run CreatePool assessment example
	@echo "🏊 Running CreatePool assessment example..."
	@go run examples/assessment/02_create_pool.go

run-assessment-swap: ## Run Swap assessment example
	@echo "💱 Running Swap assessment example..."
	@go run examples/assessment/03_swap.go

run-assessment-quote: ## Run SwapQuote assessment example
	@echo "📊 Running SwapQuote assessment example..."
	@go run examples/assessment/04_swap_quote.go

run-assessment-claim: ## Run ClaimTradingFee assessment example
	@echo "💰 Running ClaimTradingFee assessment example..."
	@go run examples/assessment/05_claim_trading_fee.go

run-assessment-withdraw: ## Run WithdrawLeftover assessment example
	@echo "🏦 Running WithdrawLeftover assessment example..."
	@go run examples/assessment/06_withdraw_leftover.go

run-assessment-damm-v1: ## Run DAMM V1 migration assessment example
	@echo "🔄 Running DAMM V1 migration assessment example..."
	@go run examples/assessment/07_migrate_damm_v1.go

run-assessment-damm-v2: ## Run DAMM V2 migration assessment example
	@echo "🔄 Running DAMM V2 migration assessment example..."
	@go run examples/assessment/08_migrate_damm_v2.go

run-assessment-all: run-assessment-config run-assessment-pool run-assessment-swap run-assessment-quote run-assessment-claim run-assessment-withdraw run-assessment-damm-v1 run-assessment-damm-v2 ## Run all assessment examples

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	@go test ./pkg/dbc -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out

bench: ## Run benchmarks
	@echo "⚡ Running benchmarks..."
	@go test ./pkg/dbc -bench=. -benchmem

# Building
build: ## Build the SDK
	@echo "🏗️  Building SDK..."
	@go build ./pkg/dbc

build-examples: ## Build example programs
	@echo "🏗️  Building examples..."
	@go build -o bin/basic_usage ./examples/basic_usage.go
	@go build -o bin/migration_example ./examples/migration_example.go
	@go build -o bin/comprehensive_test ./examples/comprehensive_test.go

build-all: build build-examples ## Build SDK and examples

# Cross-compilation
build-linux: ## Build for Linux
	@echo "🐧 Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o bin/dbc-sdk-linux ./pkg/dbc

build-macos: ## Build for macOS
	@echo "🍎 Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -o bin/dbc-sdk-macos ./pkg/dbc

build-windows: ## Build for Windows
	@echo "🪟 Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o bin/dbc-sdk-windows.exe ./pkg/dbc

build-cross: build-linux build-macos build-windows ## Cross-compile for all platforms

# Documentation
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	@godoc -http=:6060 &
	@echo "Documentation server started at http://localhost:6060"

docs-md: ## Generate markdown documentation
	@echo "📝 Generating markdown documentation..."
	@go doc -all ./pkg/dbc > docs/API.md

# Utilities
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf logs/
	@rm -f coverage.out coverage.html
	@go clean

reset: clean ## Reset environment (clean + remove test keys)
	@echo "🔄 Resetting environment..."
	@rm -rf test_keys/
	@rm -f .env

update: ## Update dependencies
	@echo "⬆️  Updating dependencies..."
	@go get -u ./...
	@go mod tidy

check: format lint vet test ## Run all checks (format, lint, vet, test)

# Examples
run-basic: ## Run basic usage example
	@echo "▶️  Running basic usage example..."
	@go run examples/basic/main.go

run-migration: ## Run migration example
	@echo "▶️  Running migration example..."
	@go run examples/migration/main.go

run-production: ## Run production example
	@echo "▶️  Running production example..."
	@go run examples/production/main.go

run-testnet: ## Run testnet integration
	@echo "▶️  Running testnet integration..."
	@go run examples/testnet_integration/main.go

# Release preparation
pre-commit: format lint vet test ## Run pre-commit checks
	@echo "✅ Pre-commit checks completed successfully!"

release-check: clean install test-coverage lint security build-cross ## Full release preparation check
	@echo "🚀 Release checks completed successfully!"

# Development workflow
dev: setup install-tools ## Set up complete development environment
	@echo "✅ Development environment ready!"

ci: install lint vet test ## CI/CD pipeline
	@echo "✅ CI pipeline completed successfully!"