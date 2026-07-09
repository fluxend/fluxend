# These are core command also available in binary CLI

server: ## Run the project in development mode
	@go run cmd/main.go server

routes.list: ## Show all the available routes
	@go run cmd/main.go routes

seed: ## Seed the database
	@go run cmd/main.go seed

about: ## Show the project information
	@go run cmd/main.go about

optimize: ## Optimize the project
	@go run cmd/main.go optimize

udb.stats: ## Show the database stats for given project
	@go run cmd/main.go udb.stats