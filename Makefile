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

.PHONY: test
download:
	@echo Download go.mod dependencies
	@go mod download
