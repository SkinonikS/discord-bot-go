GO := go
BINARY_NAME := bot
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

MAIN_GO ?= cmd/bot/main.go
TAG ?= latest
GOOS ?= linux
GOARCH ?= amd64

build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -ldflags "\
		-X 'main.tag=$(TAG)' \
		-X 'main.buildTime=$(BUILD_TIME)' \
		-X 'main.commit=$(GIT_COMMIT)'" -o $(BINARY_NAME) $(MAIN_GO)

.PHONY: build test

test:
	$(GO) test -v -race -count=1 ./...

migration-create:
	@read -p "Enter migration name: " name; \
	$(GO) run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations create $$name sql