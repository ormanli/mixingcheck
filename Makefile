GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

.PHONY: build
build:
	$(GOBUILD) -o mixingcheck -v main.go

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic -tags=integration ./...

# Go releases supported upstream right now, oldest first (e.g. "go1.26.11 go1.27.4").
GO_VERSIONS=$(shell curl -fsSL 'https://go.dev/dl/?mode=json' | grep -o '"version": *"go[0-9.]*"' | grep -o 'go[0-9.]*' | sort -uV)

# Runs the tests under every supported Go, then checks that a binary built with
# the newest Go (what a release ships) analyzes this repo under each of them.
# Exit code 3 from the analyzer means "diagnostics reported", which is expected here.
.PHONY: test-all
test-all:
	@set -e; versions="$(GO_VERSIONS)"; test -n "$$versions" || { echo "could not fetch Go versions"; exit 1; }; \
	newest=$${versions##* }; \
	GOTOOLCHAIN=$$newest $(GOBUILD) -o mixingcheck main.go; \
	for v in $$versions; do \
		echo "== $$v: tests"; GOTOOLCHAIN=$$v $(GOTEST) -race -tags=integration ./...; \
		echo "== $$v: binary built with $$newest"; GOTOOLCHAIN=$$v ./mixingcheck ./... || [ $$? -eq 3 ]; \
	done

# Moves go.mod to the oldest supported Go, updates all dependencies, then runs test-all.
.PHONY: upgrade
upgrade:
	@set -e; versions="$(GO_VERSIONS)"; test -n "$$versions" || { echo "could not fetch Go versions"; exit 1; }; \
	oldest=$${versions%% *}; oldest=$${oldest#go}; \
	$(GOCMD) mod edit -go=$${oldest%.*}.0 -toolchain=none; \
	$(GOCMD) get -u ./...; $(GOCMD) mod tidy
	@$(MAKE) test-all

.PHONY: download
download:
	@echo Download go.mod dependencies
	@go mod download
