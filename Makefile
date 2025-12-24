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
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":"}; {target=$$2; sub(/^[[:space:]]*/, "", target); desc=$$3; for(i=4; i<=NF; i++) desc=desc":"$$i; sub(/.*## /, "", desc); printf "\033[36m%-30s\033[0m %s\n", target, desc}'


# Include standard makes
include ./scripts/makes/core.mk


# Docker targets
.PHONY: docker-up-local
docker-up-local: ## Start Docker services, be sure to setup /etc/hosts entry for local-techboss
	sudo ifconfig lo0 alias 127.10.0.1/8
	ifconfig lo0 | grep 127.10.0.1
	docker compose -f infra/docker-compose-local.yml up


.PHONY: claude
claude: ## Create Claude PR - Usage: make claude BRANCH=feature-name TASK="description"
	@if [ -z "$(TASK)" ]; then \
		echo "❌ Error: TASK is required"; \
		echo "Usage: make claude BRANCH=my-feature TASK=\"Add user authentication\""; \
		exit 1; \
	fi; \
	BASE_BRANCH=$${BRANCH:-development}; \
	./scripts/claude-pr.sh "$(TASK)" "$$BASE_BRANCH"
	