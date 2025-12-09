# Development targets
.PHONY: server
server: ## Run the server
	NO_MIGRATE="true" SYS_ENV=$(SYS_ENV) CONFIG_FILE=$(CONFIG_FILE) REGION=$(REGION) go run ./cmd/server

.PHONY: dev
dev: ## Run the server with hot-reload using gow
	air server

.PHONY: migrate
migrate: ## Run database migrations
	SYS_ENV=$(SYS_ENV) CONFIG_FILE=$(CONFIG_FILE) REGION=$(REGION) go run ./cmd/migrations

# Docker targets
.PHONY: docker-up
docker-up: ## Start Docker services, be sure to setup /etc/hosts entry for local-techboss
	docker compose -f infra/docker-compose-local.yml up


# Code quality targets
.PHONY: lint
PKG ?= ./...
lint: ## Run lint (Usage: make lint PKG=./internal/services/evidencing)
	@bash -cex '\
		golangci-lint run $(PKG)\
	'

.PHONY: fmt
PKG ?= ./...
fmt: ## Run fmt (Usage: make fmt PKG=./internal/services/evidencing)
	@bash -cex '\
		golangci-lint fmt $(PKG)\
	'

# Code quality targets
.PHONY: lint-fast
lint-fast: ## Run golangci-lint
	golangci-lint run --fast-only



# Dependency management
.PHONY: update-deps
update-deps: ## Update all Go modules
	go list -m -u all | awk '{print $$1}' | xargs -n 1 go get -u
	go mod tidy

.PHONY: tidy
tidy: ## Tidy up Go modules
	@bash -c '\
		export GOPRIVATE=github.com/griffnb/core/*; \
		export GH_TOKEN=$$(gh auth token); \
		git config --global url."https://$${GH_TOKEN}@github.com/".insteadOf "https://github.com/"; \
		go mod tidy; \
	'


.PHONY: test
PKG ?= ./...
RUN ?=
EXTRA ?=

test: ## Run tests (Usage: make test PKG=./internal/services/evidencing RUN='TestName/Subtest')
	@bash -cex '\
		SYS_ENV="unit_test" \
		CONFIG_FILE="$$(pwd)/.configs/unit_test.json" \
		REGION="us-east-1" \
		go test $(PKG) -v -count=1 -timeout=30s -race $(if $(RUN),-run "^$(RUN)$$",) $(EXTRA) \
	'





.PHONY: runner
runner: ## Run matcher wizard with dynamic args
	SYS_ENV=$(SYS_ENV) CONFIG_FILE=$(CONFIG_FILE) REGION=$(REGION) \
		go run ./cmd/runner $(ARGS)

.PHONY: runner-stage
runner-stage: ## Run matcher wizard to interactively select broker, files, and options
	@bash -c '\
		./scripts/login_stage.sh; \
		CLOUD=true SYS_ENV=$(SYS_ENV) CONFIG_FILE=$(STAGE_CONFIG_FILE) REGION=$(REGION) go run ./cmd/runner $(ARGS) \
	'

.PHONY: runner-prod
runner-prod: ## Run matcher wizard to interactively select broker, files, and options
	@bash -c '\
		./scripts/login_prod.sh; \
		CLOUD=true SYS_ENV=$(SYS_ENV) CONFIG_FILE=$(PROD_CONFIG_FILE) REGION=$(REGION) go run ./cmd/runner $(ARGS) \
	'




