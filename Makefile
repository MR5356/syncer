NAME = syncer
BASE_VERSION = build
MODULE_NAME = github.com/MR5356/syncer

VERSION ?= $(shell echo "${BASE_VERSION}.")$(shell git rev-parse --short HEAD)

.DEFAULT_GOAL := help

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

build: clean deps ## Build the project
	go build -ldflags "-s -w -X '$(MODULE_NAME)/pkg/version.Version=$(VERSION)'" -o build/$(NAME) ./cmd/syncer

all: release ## Generate releases for all supported systems

release: clean deps ## Generate releases for unix systems
	chmod +x hack/release.sh
	bash -c "hack/release.sh $(VERSION) $(NAME) $(MODULE_NAME)"

test: deps ## Execute tests
	go test ./...

deps: ## Install dependencies using go get
	go get -d -v -t ./...

clean: ## Remove building artifacts
	rm -rf build
	rm -f $(NAME)
