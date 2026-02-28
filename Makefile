GO_BINARY := go
BOT_BINARY_NAME := bot
CLI_BINARY_NAME := cli
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

BOT_MAIN_GO ?= cmd/bot/main.go
CLI_MAIN_GO ?= cmd/cli/main.go
TAG ?= latest
GOOS ?= linux
GOARCH ?= amd64

.PHONY: build test

build-bot:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO_BINARY) build -ldflags "\
		-X 'main.tag=$(TAG)' \
		-X 'main.buildTime=$(BUILD_TIME)' \
		-X 'main.commit=$(GIT_COMMIT)'" -o $(BOT_BINARY_NAME) $(BOT_MAIN_GO)

build-cli:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO_BINARY) build -ldflags "\
		-X 'main.tag=$(TAG)' \
		-X 'main.buildTime=$(BUILD_TIME)' \
		-X 'main.commit=$(GIT_COMMIT)'" -o $(CLI_BINARY_NAME) $(CLI_MAIN_GO)

test:
	$(GO_BINARY) test -v -race -count=1 ./...

migration-create:
	@read -p "Enter migration name: " name; \
	$(GO_BINARY) run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations create $$name sql