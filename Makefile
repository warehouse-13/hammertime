BIN_DIR := bin
HT_CMD := .

BUILD_DATE := $(shell date +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --tags --abbrev=0)
VERSION_PKG := github.com/warehouse-13/hammertime/pkg/version

##@ Build

.PHONY: build
build: ## Build hammertime
	go build -o hammertime main.go

.PHONY: lint
lint: ## Lint code
	golangci-lint run -v --fast=false

##@ Test

.PHONY: test
test: int unit

.PHONY: int
int: ## Run integration tests (ginkgo)
	ginkgo -r test/

.PHONY: unit
unit: ## Run unit tests
	go test ./pkg/...

.PHONY: mock
mock: ## Generate mocks
	go generate ./...

##@ Release

.PHONY: release
release: ## Cross compile bins for linux, windows, mac
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/hammertime-linux-amd64 -ldflags "-X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).BuildDate=$(BUILD_DATE) -X $(VERSION_PKG).CommitHash=$(GIT_COMMIT)" $(HT_CMD)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(BIN_DIR)/hammertime-linux-arm64 -ldflags "-X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).BuildDate=$(BUILD_DATE) -X $(VERSION_PKG).CommitHash=$(GIT_COMMIT)" $(HT_CMD)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/hammertime-windows-amd64 -ldflags "-X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).BuildDate=$(BUILD_DATE) -X $(VERSION_PKG).CommitHash=$(GIT_COMMIT)" $(HT_CMD)
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o $(BIN_DIR)/hammertime-windows-arm64 -ldflags "-X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).BuildDate=$(BUILD_DATE) -X $(VERSION_PKG).CommitHash=$(GIT_COMMIT)" $(HT_CMD)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/hammertime-darwin-amd64 -ldflags "-X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).BuildDate=$(BUILD_DATE) -X $(VERSION_PKG).CommitHash=$(GIT_COMMIT)" $(HT_CMD)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/hammertime-darwin-arm64 -ldflags "-X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).BuildDate=$(BUILD_DATE) -X $(VERSION_PKG).CommitHash=$(GIT_COMMIT)" $(HT_CMD)

.PHONY: help
help:  ## Display this help. Thanks to https://www.thapaliya.com/en/writings/well-documented-makefiles/
ifeq ($(OS),Windows_NT)
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  %-40s %s\n", $$1, $$2 } /^##@/ { printf "\n%s\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
else
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-40s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
endif
