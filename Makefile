BIN_DIR := bin
HT_CMD := .

.PHONY: build
build: ## Build hammertime
	go build -o hammertime main.go

.PHONY: test
test: int unit
	
.PHONY: int
int: ## Run integration tests
		ginkgo -r test/

.PHONY: unit
unit: ## Run unit tests
		ginkgo -r pkg/

.PHONY: release
release: ## Cross compile bins for linux, windows, mac
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/hammertime-linux-amd64 $(HT_CMD)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(BIN_DIR)/hammertime-linux-arm64 $(HT_CMD)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/hammertime-windows-amd64 $(HT_CMD)
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o $(BIN_DIR)/hammertime-windows-arm64 $(HT_CMD)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/hammertime-darwin-amd64 $(HT_CMD)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/hammertime-darwin-arm64 $(HT_CMD)

.PHONY: help
help:  ## Display this help. Thanks to https://www.thapaliya.com/en/writings/well-documented-makefiles/
ifeq ($(OS),Windows_NT)
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  %-40s %s\n", $$1, $$2 } /^##@/ { printf "\n%s\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
else
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-40s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
endif
