lint: ## Run linter
	@golangci-lint run

lint.fix: ## Run linter and fix
	@golangci-lint run --fix

test: ## Run tests
	@go test -v ./... -coverprofile=coverage.out

