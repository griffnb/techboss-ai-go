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


.PHONY: code-gen-ts
code-gen-ts: ## Create TypeScript models (Usage: make code-gen-ts ModelName [package_name=package])
	@bash -c '\
		MODEL_NAME="$(filter-out $@,$(MAKECMDGOALS))"; \
		PACKAGE_NAME=$$(grep "^module " go.mod | sed "s/module //"); \
		if [ -z "$$MODEL_NAME" ]; then \
			echo "Error: Model name required. Usage: make code-gen-ts ModelName"; \
			exit 1; \
		fi; \
		PKG="$(package_name)"; \
		if [ -z "$$PKG" ]; then \
			PKG=""; \
		fi; \
		GOPACKAGE=$$PACKAGE_NAME core_gen typescript "$$MODEL_NAME" "-modelPackage=$$PKG"; \
	'
%:
	@: