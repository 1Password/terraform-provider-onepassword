export MAIN_BRANCH ?= main

.DEFAULT_GOAL := help
.PHONY: test testacc build clean test/coverage release/prepare release/tag .check_bump_type .check_git_clean help test-e2e test-e2e-service-account test-e2e-connect

GIT_BRANCH := $(shell git symbolic-ref --short HEAD 2>/dev/null || echo "")
WORKTREE_CLEAN := $(shell git status --porcelain 1>/dev/null 2>&1; echo $$?)
SCRIPTS_DIR := $(CURDIR)/scripts

versionFile = $(CURDIR)/.VERSION

curVersion := $(shell cat $(versionFile) | sed 's/^v//')

test:	## Runs integration and unit tests
	TF_ACC=1 go test $(shell go list ./... | grep -v /test/e2e)

test/coverage:	## Runs integration and unit tests with coverage report
	TF_ACC=1 go test $(shell go list ./... | grep -v /test/e2e)

testacc: ## Run acceptance tests
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

test-e2e: test-e2e-service-account test-e2e-connect ## Run all e2e tests (service account and Connect)

test-e2e-service-account: ## Run e2e tests using service account (requires OP_SERVICE_ACCOUNT_TOKEN)
	@test -n "$(OP_SERVICE_ACCOUNT_TOKEN)" || (echo "[ERROR] OP_SERVICE_ACCOUNT_TOKEN environment variable is not set."; exit 1)
	@echo "[INFO] Running e2e tests with service account authentication..."
	@sh -c 'unset OP_CONNECT_TOKEN OP_CONNECT_HOST; OP_SERVICE_ACCOUNT_TOKEN="$(OP_SERVICE_ACCOUNT_TOKEN)" TF_ACC=1 go test -v ./test/e2e/... -timeout 30m'

test-e2e-connect: ## Run e2e tests using Connect (requires OP_CONNECT_TOKEN and OP_CONNECT_HOST)
	@test -n "$(OP_CONNECT_TOKEN)" || (echo "[ERROR] OP_CONNECT_TOKEN environment variable is not set."; exit 1)
	@test -n "$(OP_CONNECT_HOST)" || (echo "[ERROR] OP_CONNECT_HOST environment variable is not set."; exit 1)
	@echo "[INFO] Running e2e tests with Connect authentication..."
	@sh -c 'unset OP_SERVICE_ACCOUNT_TOKEN; OP_CONNECT_TOKEN="$(OP_CONNECT_TOKEN)" OP_CONNECT_HOST="$(OP_CONNECT_HOST)" TF_ACC=1 go test -v ./test/e2e/... -timeout 30m'

build: clean	## Build project
	go build -o ./dist/terraform-provider-onepassword .

clean:
	rm -rf ./dist

help:	## Prints this help message
	@grep -E '^[\/a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


## Release functions =====================

release/prepare: .check_git_clean	## Updates changelog and creates release branch (call with 'release/prepare version=<new_version_number>')

	@test $(version) || (echo "[ERROR] version argument not set."; exit 1)
	@git fetch --quiet origin $(MAIN_BRANCH)

	@echo $(version) | tr -d '\n' | tee $(versionFile) &>/dev/null

	@NEW_VERSION=$(version) $(SCRIPTS_DIR)/prepare-release.sh

release/tag: .check_git_clean	## Creates git tag
	@git pull --ff-only
	@echo "Applying tag 'v$(curVersion)' to HEAD..."
	@git tag --sign "v$(curVersion)" -m "Release v$(curVersion)"
	@echo "[OK] Success!"
	@echo "Remember to call 'git push --tags' to persist the tag."

## Helper functions =====================

.check_git_clean:
ifneq ($(GIT_BRANCH), $(MAIN_BRANCH))
	@echo "[ERROR] Please checkout default branch '$(MAIN_BRANCH)' and re-run this command."; exit 1;
endif
ifneq ($(WORKTREE_CLEAN), 0)
	@echo "[ERROR] Uncommitted changes found in worktree. Address them and try again."; exit 1;
endif
