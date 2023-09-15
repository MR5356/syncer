NAME = syncer
BASE_VERSION = v1.0.0-release_build
MODULE_NAME = github.com/MR5356/syncer

VERSION ?= $(shell echo "${BASE_VERSION}.")$(shell git rev-parse --short HEAD)

OS = linux darwin windows
architecture = amd64 arm64

.DEFAULT_GOAL := help

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

build: clean deps ## Build the project
	go build -ldflags "-s -w -X '$(MODULE_NAME)/pkg/version.Version=$(VERSION)'" -o build/$(NAME) ./cmd/syncer

all: release ## Generate releases for all supported systems

release: clean deps ## Generate releases for unix systems
	@for arch in $(architecture);\
	do \
		for os in $(OS);\
		do \
			echo "Building $$os-$$arch"; \
			mkdir -p build/$$os-$$arch; \
			if [ "$$os" == "windows" ]; then \
				GOOS=$$os GOARCH=$$arch go build -ldflags "-s -w -X '$(MODULE_NAME)/pkg/version.Version=$(VERSION)'" -o build/$$os-$$arch/$(NAME).exe ./cmd/syncer; \
			else \
				GOOS=$$os GOARCH=$$arch go build -ldflags "-s -w -X '$(MODULE_NAME)/pkg/version.Version=$(VERSION)'" -o build/$$os-$$arch/$(NAME) ./cmd/syncer; \
			fi; \
			cd build; \
			tar zcvf $(NAME)-$$os-$$arch.tar.gz $$os-$$arch; \
			rm -rf $$os-$$arch; \
			cd ..; \
		done \
	done

test: deps ## Execute tests
	go test ./...

deps: ## Install dependencies using go get
	go get -d -v -t ./...

clean: ## Remove building artifacts
	rm -rf build
	rm -f $(NAME)

version: ## Print version
	@echo $(VERSION)