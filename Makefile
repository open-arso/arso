CLI_DIR := apps/cli
BIN_DIR := bin
BIN_NAME := arso

VERSION := $(shell cat VERSION 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := \
	-X github.com/openarso/arso/apps/cli/cmd.Version=$(VERSION) \
	-X github.com/openarso/arso/apps/cli/cmd.Commit=$(COMMIT) \
	-X github.com/openarso/arso/apps/cli/cmd.Date=$(DATE)

.PHONY: cli-run
cli-run:
	cd $(CLI_DIR) && go run -ldflags "$(LDFLAGS)" . version

.PHONY: cli-build
cli-build:
	mkdir -p $(BIN_DIR)
	cd $(CLI_DIR) && go build -ldflags "$(LDFLAGS)" -o ../../$(BIN_DIR)/$(BIN_NAME) .

.PHONY: cli-test
cli-test:
	cd $(CLI_DIR) && go test ./...

.PHONY: cli-fmt
cli-fmt:
	cd $(CLI_DIR) && go fmt ./...

.PHONY: cli-tidy
cli-tidy:
	cd $(CLI_DIR) && go mod tidy