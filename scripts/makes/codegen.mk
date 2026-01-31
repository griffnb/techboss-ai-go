# Code generation
.PHONY: code-gen
code-gen: ## Create a new object (Usage: make create-object WORD=objectname)
	@bash -c '\
		export GOPRIVATE=github.com/griffnb/core/*; \
		export GH_TOKEN=$$(gh auth token); \
		go install github.com/griffnb/core/core_gen@latest; \
		./scripts/code_gen.sh; \
	'

.PHONY: code-gen-object
code-gen-object: ## Create a new object (Usage: make code-gen-object model_name=object_name model_plural=object_plural)
	@bash -c '\
		export GOPRIVATE=github.com/griffnb/core/*; \
		export GH_TOKEN=$$(gh auth token); \
		go install github.com/griffnb/core/core_gen@latest; \
		PACKAGE_NAME=$$(grep "^module " go.mod | sed "s/module //"); \
		core_gen object "$(model_name)" "-plural=$(model_plural)" "-modelPackage=$${PACKAGE_NAME}"; \
		go generate "./internal/models/$(model_name)/$(model_name).go"; \
		go generate "./internal/controllers/$(model_plural)/setup.go"; \
	'
.PHONY: code-gen-public-object
code-gen-public-object: ## Create a new public object (Usage: make code-gen-object model_name=object_name model_plural=object_plural)
	@bash -c '\
		export GOPRIVATE=github.com/griffnb/core/*; \
		export GH_TOKEN=$$(gh auth token); \
		go install github.com/griffnb/core/core_gen@latest; \
		PACKAGE_NAME=$$(grep "^module " go.mod | sed "s/module //"); \
		core_gen object "$(model_name)" "-plural=$(model_plural)" "-modelPackage=$${PACKAGE_NAME}" "-public=true"; \
		go generate "./internal/models/$(model_name)/$(model_name).go"; \
		go generate "./internal/controllers/$(model_plural)/setup.go"; \
	'

.PHONY: generate
generate: ## Installs latest, then runs code generation
	@bash -c '\
		export GOPRIVATE=github.com/griffnb/core/*; \
		export GH_TOKEN=$$(gh auth token); \
		go install github.com/griffnb/core/core_gen@latest; \
		files=$$(find . -name "*.go" -exec grep -l "//go:generate" {} \;); \
		echo "$$files" | xargs -n1 -P5 go generate; \
	'

	

.PHONY: generate-only
generate-only: ## Run code generation only without installing
	@bash -c '\
		files=$$(find . -name "*.go" -exec grep -l "//go:generate" {} \;); \
		echo "$$files" | xargs -n1 -P5 go generate; \
	'


.PHONY: config-schema
config-schema: ## Generate JSON schema for config files
	go run internal/environment/schema/generate-config-schema.go > internal/environment/schema/config-schema.json
	@echo "Generated config-schema.json"


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

.PHONY: code-gen-constants
code-gen-constants: ## Create TypeScript constants (Usage: make code-gen-constants package)
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
		GOPACKAGE=$$PACKAGE_NAME core_gen typescript "$$MODEL_NAME" "-modelPackage=$$PKG" -constants; \
	'
%:
	@:




# Capture arguments after target for claude/ralph commands
# Usage: make claude "do this task" or make ralph "implement feature"
ifneq (,$(filter code-claude code-ralph,$(firstword $(MAKECMDGOALS))))
  TASK := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  $(eval $(TASK):;@:)
endif

.PHONY: code-claude
code-claude: ## Create Claude PR - Usage: make code-claude "description"
	@if [ -z "$(TASK)" ]; then \
		echo "❌ Error: TASK is required"; \
		echo "Usage: make claude \"Add user authentication\""; \
		exit 1; \
	fi; \
	BASE_BRANCH=$${BRANCH:-$$(git rev-parse --abbrev-ref HEAD)}; \
	./scripts/claude-pr.sh "$(TASK)" "$$BASE_BRANCH"


.PHONY: code-ralph
code-ralph: ## Create Ralph PR - Usage: make code-ralph "description"
	@if [ -z "$(TASK)" ]; then \
		echo "❌ Error: TASK is required"; \
		echo "Usage: make ralph \"Add user authentication\""; \
		exit 1; \
	fi; \
	BASE_BRANCH=$${BRANCH:-$$(git rev-parse --abbrev-ref HEAD)}; \
	./scripts/ralph-pr.sh "$(TASK)" "$$BASE_BRANCH"