PROJECT_NAME := "jettison"
PKG := "github.com/luno/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

.PHONY: vet fmt checkfmt test race

vet: ## Lint the files
	@go vet ${PKG_LIST}

fmt: ## Format the files
	@gofumpt -w .

checkfmt: ## Check that files are formatted
	@./checkfmt.sh

test: ## Run unittests
	@go test -short ${PKG_LIST}

race: ## Run data race detector
	@go test -race -short ${PKG_LIST}

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
