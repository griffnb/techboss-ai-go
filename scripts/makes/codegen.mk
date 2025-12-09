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


.PHONY: ts-gen
ts-gen: ## Run TypeScript operations on a table (Usage: make ts-gen TABLE=tablename)
	SYS_ENV=$(SYS_ENV) CONFIG_FILE=$(CONFIG_FILE) REGION=$(REGION) go run ./cmd/ts --table="$(TABLE)"


.PHONY: config-schema
config-schema: ## Generate JSON schema for config files
	go run internal/environment/schema/generate-config-schema.go > internal/environment/schema/config-schema.json
	@echo "Generated config-schema.json"