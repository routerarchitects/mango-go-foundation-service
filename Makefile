# ==============================================================================
# Development Makefile for {{SERVICE_NAME}}
# ==============================================================================

APP_NAME = mango-go-foundation-service
VERSION ?= v0.1.0
COMMIT_HASH = $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIMESTAMP = $(shell date -u +%s)
LDFLAGS = -X github.com/routerarchitects/ra-common-mods/buildinfo.version=$(VERSION) \
          -X github.com/routerarchitects/ra-common-mods/buildinfo.buildTimestamp=$(BUILD_TIMESTAMP) \
          -X github.com/routerarchitects/ra-common-mods/buildinfo.commitHash=$(COMMIT_HASH)

.PHONY: all build run test tidy fmt lint clean docker-build docker-run

all: build

build:
	@echo "Compiling {{SERVICE_NAME}} binary..."
	@mkdir -p bin
	go build -ldflags="-s -w $(LDFLAGS)" -o bin/$(APP_NAME) ./cmd

run:
	@echo "Running {{SERVICE_NAME}} locally..."
	@if [ -f env/local-dev.env ]; then \
		set -a && . ./env/local-dev.env && set +a && go run ./cmd; \
	else \
		go run ./cmd; \
	fi

test:
	@echo "Running unit tests..."
	go test -v -race ./...

tidy:
	@echo "Tidying module dependencies..."
	go mod tidy

fmt:
	@echo "Formatting source files..."
	go fmt ./...

lint:
	@echo "Analyzing source code..."
	go vet ./...

clean:
	@echo "Cleaning build outputs..."
	rm -rf bin/

docker-build:
	@echo "Building Docker image $(APP_NAME):latest..."
	docker build \
		--build-arg APP_NAME=$(APP_NAME) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP) \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		-t $(APP_NAME):latest .

docker-run:
	@echo "Starting Docker container $(APP_NAME) in foreground..."
	docker run --rm -it \
		--env-file env/docker-compose.env \
		-v $(PWD)/certs:/app/certs \
		-p 8088:8088 -p 17007:17007 \
		$(APP_NAME):latest
