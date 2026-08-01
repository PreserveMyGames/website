APP_NAME := preservemygames-store
BIN_DIR := bin
CMD := ./cmd/store
VERSION ?= $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/PreserveMyGames/store/internal/web.assetVersion=$(VERSION)

.PHONY: build run test vet fmt tidy ci docker-build

build:
	go build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) $(CMD)

run: build
	DATA_DIR=./data ./$(BIN_DIR)/$(APP_NAME)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w $$(go list -f '{{.Dir}}' ./...)

tidy:
	go mod tidy

ci: fmt vet test build

docker-build:
	docker build -f docker/Dockerfile -t preservemygames/store:latest .
