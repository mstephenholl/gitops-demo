.PHONY: build test test-cover lint lint-fix docker-build k3d-setup k3d-teardown k3d-import run clean

# ---- Load .env (if present) ----
-include .env
export

# ---- Variables ----
APP_NAME   ?= gitops-demo
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
MODULE     := github.com/mstephenholl/gitops-demo

LDFLAGS := -s -w \
  -X $(MODULE)/internal/version.Tag=$(VERSION) \
  -X $(MODULE)/internal/version.Commit=$(COMMIT) \
  -X $(MODULE)/internal/version.BuildTime=$(BUILD_TIME)

# ---- Build ----
build:
	go build -ldflags "$(LDFLAGS)" -o bin/server ./cmd/server

run: build
	./bin/server

# ---- Test ----
test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "Open HTML report: go tool cover -html=coverage.out"

# ---- Lint ----
lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

# ---- Docker ----
docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(APP_NAME):latest .

# ---- k3d ----
k3d-setup:
	chmod +x scripts/setup.sh && ./scripts/setup.sh

k3d-teardown:
	chmod +x scripts/teardown.sh && ./scripts/teardown.sh

k3d-import: docker-build
	k3d image import $(APP_NAME):latest -c $(APP_NAME)

# ---- Clean ----
clean:
	rm -rf bin/ coverage.out
