.PHONY: rebuild-unit-test-db
rebuild-unit-test-db: ## Rebuild the unit test database
	SYS_ENV="unit_test" CONFIG_FILE="$$(pwd)/.configs/unit_test.json" REGION="us-east-1" go run cmd/scripts/rebuild_unit_database.go



.PHONY: deadcode
deadcode: ## Run deadcode to find unused code
	deadcode ./cmd/server

.PHONY: swagger-docs
swagger-docs: ## Swagger docs generation script, # generates swagger docs, install the forked version of swaggo from griffnb then do go install ./cmd/swag, then run this script
	@swag init -g "main.go" -d "./cmd/server,./internal/controllers,./internal/models" --parseInternal -pd -o "./swag_docs"

.PHONY: swagger
swagger: ## Serve swagger docs at http://localhost:1323/swagger/index.html
	@swag init -g "main.go" -d "./cmd/server,./internal/controllers,./internal/models" --parseInternal -pd -o "./swag_docs"
	@echo "Serving swagger docs at http://localhost:1323/swagger/index.html"
	@echo "Press Ctrl+C to stop"
	@open -a "Google Chrome" http://localhost:1323/swagger/index.html
	@SYS_ENV=$(SYS_ENV) CONFIG_FILE=$(CONFIG_FILE) REGION=$(REGION) go run ./cmd/swagger 

.PHONY: swagger-format
swagger-format: ## Format swagger docs
	@swag fmt -d "./cmd/server,./internal/controllers,./internal/models"





