# Techboss AI Go Project Makefile
# ---------------------------

# Default shell for executing commands
SHELL := /bin/bash

# Environment variables
SYS_ENV ?= development
CONFIG_FILE ?= ./.configs/development.json
PROD_CONFIG_FILE ?= ./.configs/production.json
STAGE_CONFIG_FILE ?= ./.configs/staging.json
REGION ?= us-east-1

 
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development targets
.PHONY: server
server: ## Run the server
	NO_MIGRATE="true" SYS_ENV=$(SYS_ENV) CONFIG_FILE=$(CONFIG_FILE) REGION=$(REGION) go run ./cmd/server

.PHONY: dev
dev: ## Run the server with hot-reload using gow
	NO_MIGRATE="true" SYS_ENV=$(SYS_ENV) CONFIG_FILE=$(CONFIG_FILE) REGION=$(REGION) gow run ./cmd/server

.PHONY: migrate
migrate: ## Run database migrations
	SYS_ENV=$(SYS_ENV) CONFIG_FILE=$(CONFIG_FILE) REGION=$(REGION) go run ./cmd/migrations

# Docker targets
.PHONY: docker-up
docker-up: ## Start Docker services
	docker compose -f infra/docker-compose.yml up

# Code quality targets
.PHONY: lint
PKG ?= ./...
lint: # Run lint PKG=./internal/services/evidencing
	@bash -cex '\
		golangci-lint run $(PKG)\
	'
%:
	@:

.PHONY: fmt
PKG ?= ./...
fmt: # Run fmt PKG=./internal/services/evidencing
	@bash -cex '\
		golangci-lint fmt $(PKG)\
	'
%:
	@:

# Code quality targets
.PHONY: lint-fast
lint-fast: ## Run golangci-lint
	golangci-lint run --fast-only


.PHONY: install-lint
install-lint: ## Install golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	$$(go env GOPATH)/bin/golangci-lint --version


.PHONY: install-gow
install-gow: ## Install gow
	go install github.com/mitranim/gow@latest
	gow version

# Dependency management
.PHONY: update-deps
update-deps: ## Update all Go modules
	go list -m -u all | awk '{print $$1}' | xargs -n 1 go get -u
	go mod tidy

.PHONY: tidy
tidy: ## Tidy up Go modules
	@bash -c '\
		export GOPRIVATE=github.com/CrowdShield/*; \
		export GH_TOKEN=$$(gh auth token); \
		git config --global url."https://$${GH_TOKEN}@github.com/".insteadOf "https://github.com/"; \
		go mod tidy; \
	'

.PHONY: test
PKG ?= ./...
RUN ?=
EXTRA ?=

test: ## Run tests make test PKG=./internal/services/evidencing RUN='TestName/Subtest'
	@bash -cex '\
		SYS_ENV="unit_test" \
		CONFIG_FILE="$$(pwd)/.configs/unit_test.json" \
		REGION="us-east-1" \
		go test $(PKG) -v -count=1 -timeout=30s -race $(if $(RUN),-run "^$(RUN)$$",) $(EXTRA) \
	'

%:
	@:

.PHONY: install-private
install-private: ## Install private Go modules
	@bash -c '\
		export GOPRIVATE=github.com/CrowdShield/*; \
		export GH_TOKEN=$$(gh auth token); \
		go get $(filter-out $@,$(MAKECMDGOALS)); \
	'
# Prevent Make from interpreting the args as targets
%:
	@:

# Code generation
.PHONY: code-gen
code-gen: ## Create a new object (Usage: make create-object WORD=objectname)
	@bash -c '\
		export GOPRIVATE=github.com/CrowdShield/*; \
		export GH_TOKEN=$$(gh auth token); \
		go install github.com/CrowdShield/go-core/core_generate@latest; \
		./scripts/code_gen.sh; \
	'

.PHONY: code-gen-object
code-gen-object: ## Create a new object (Usage: make code-gen-object model_name=object_name model_plural=object_plural)
	@bash -c '\
		export GOPRIVATE=github.com/CrowdShield/*; \
		export GH_TOKEN=$$(gh auth token); \
		go install github.com/CrowdShield/go-core/core_generate@latest; \
		PACKAGE_NAME=$$(grep "^module " go.mod | sed "s/module //"); \
		core_generate object "$(model_name)" "-plural=$(model_plural)" "-modelPackage=$${PACKAGE_NAME}"; \
		go generate "./internal/models/$(model_name)/$(model_name).go"; \
		go generate "./internal/controllers/$(model_plural)/setup.go"; \
	'

.PHONY: generate
generate: ## Installs latest, then runs code generation
	@bash -c '\
		export GOPRIVATE=github.com/CrowdShield/*; \
		export GH_TOKEN=$$(gh auth token); \
		go install github.com/CrowdShield/go-core/core_generate@latest; \
		files=$$(find . -name "*.go" -exec grep -l "//go:generate" {} \;); \
		echo "$$files" | xargs -n1 -P5 go generate; \
	'

	

.PHONY: generate-only
generate-only: ## Run code generation only without installing
	@bash -c '\
		files=$$(find . -name "*.go" -exec grep -l "//go:generate" {} \;); \
		echo "$$files" | xargs -n1 -P5 go generate; \
	'

.PHONY: install-codegen
install-codegen: ## Install latest code generation from go-core
	@bash -c '\
		export GOPRIVATE=github.com/CrowdShield/*; \
		export GH_TOKEN=$$(gh auth token); \
		go install github.com/CrowdShield/go-core/core_generate@latest; \
	'

.PHONY: ts-gen
ts-gen: ## Run TypeScript operations on a table (Usage: make ts-gen TABLE=tablename)
	@if [ -z "$(TABLE)" ]; then echo "TABLE parameter is required. Usage: make ts TABLE=tablename"; exit 1; fi
	SYS_ENV=$(SYS_ENV) CONFIG_FILE=$(CONFIG_FILE) REGION=$(REGION) go run ./cmd/ts --table="$(TABLE)"


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


# Third-party services
.PHONY: stripe-login
stripe-login: ## Login to Stripe
	stripe login

.PHONY: stripe-webhook
stripe-webhook: ## Start Stripe webhook listener
	stripe listen --forward-to http://localhost:8080/billing/stripe/webhook

.PHONY: stripe-install
stripe-install: ## Install Stripe CLI
	brew install stripe/stripe-cli/stripe
	stripe login




.PHONY: hotfix
hotfix:
	@if [ -z "$(title)" ]; then \
		echo "Usage: make hotfix title='Fixes this bug'"; \
		exit 1; \
	fi; \
	current_branch=$$(git rev-parse --abbrev-ref HEAD); \
	gh pr create --base main --head $$current_branch --title "HOTFIX PROD: $(title)" --body "$(title)"; \
	gh pr create --base development --head $$current_branch --title "HOTFIX STAGE: $(title)" --body "$(title)"


.PHONY: install-gh
install-gh: ## Install GitHub CLI
	brew install gh
	gh auth login

.PHONY: gh-login
gh-login: ## Login to GitHub CLI
	gh auth login


.PHONY: install-deadcode
install-deadcode: ## Install deadcode
	go install golang.org/x/tools/cmd/deadcode@latest

.PHONY: deadcode
deadcode: ## Run deadcode to find unused code
	deadcode ./cmd/server
