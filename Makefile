.PHONY: all build run test test-coverage test-verbose lint lint-install fmt vet tidy \
        docker-build docker-up docker-down docker-logs clean

ifeq ($(OS),Windows_NT)
  BINARY       := bin/server.exe
  MKDIR        := if not exist bin mkdir bin
  RM           := if exist bin rmdir /s /q bin
  GOPATH_BIN   := $(shell go env GOPATH)\bin
  LINT_BIN     := $(GOPATH_BIN)\golangci-lint.exe
else
  BINARY       := bin/server
  MKDIR        := mkdir -p bin
  RM           := rm -rf bin coverage.out coverage.html
  GOPATH_BIN   := $(shell go env GOPATH)/bin
  LINT_BIN     := $(GOPATH_BIN)/golangci-lint
endif

CMD          := ./cmd/server
GO           := go
GOFLAGS      := -ldflags="-s -w"
COVERAGE     := coverage.out
LINT_VERSION := v1.62.2

all: tidy fmt vet test build

build:
	$(MKDIR)
	$(GO) build $(GOFLAGS) -o $(BINARY) $(CMD)

run: build
	$(BINARY)

test:
	$(GO) test -race -count=1 ./...

test-coverage:
	$(GO) test -race -count=1 -coverprofile=$(COVERAGE) -covermode=atomic ./...
	$(GO) tool cover -html=$(COVERAGE) -o coverage.html

test-verbose:
	$(GO) test -race -v -count=1 ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

lint-install:
ifeq ($(OS),Windows_NT)
	winget install golangci-lint
else
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_BIN) $(LINT_VERSION)
endif

lint:
	"$(LINT_BIN)" run ./...

docker-build:
	docker compose build

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v

docker-logs:
	docker compose logs -f

clean:
	$(RM)