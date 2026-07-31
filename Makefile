APP_NAME := preservemygames-web
BIN_DIR := bin
CMD := ./cmd/web
IMAGE := preservemygames/website:latest
DOCKERFILE := docker/Dockerfile
COMPOSE := docker compose -f docker/docker-compose.yml
GOSEC := $(shell go env GOPATH)/bin/gosec

.PHONY: build run test test-race test-fuzz bench vet fmt fix tidy vendor gosec docker-build docker-up docker-down docker-test ci

build:
	go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME) $(CMD)

run: build
	./$(BIN_DIR)/$(APP_NAME)

test:
	go test ./...

test-race:
	go test -race ./...

test-fuzz:
	go test -fuzz=FuzzSplitFrontMatter -fuzztime=5s ./internal/blog
	go test -fuzz=FuzzParsePostNoPanic -fuzztime=5s ./internal/blog
	go test -fuzz=FuzzExcerptValidUTF8 -fuzztime=5s ./internal/blog
	go test -fuzz=FuzzSlug -fuzztime=5s ./internal/validate
	go test -fuzz=FuzzStaticPath -fuzztime=5s ./internal/validate
	go test -fuzz=FuzzRedirectTarget -fuzztime=5s ./internal/web

bench:
	go test -bench=. -benchmem ./internal/web

vet:
	go vet ./...

fmt:
	gofmt -w $$(go list -f '{{.Dir}}' ./...)

fix:
	go fix ./...

tidy:
	go mod tidy

vendor:
	./scripts/vendor.sh

gosec:
	@test -x $(GOSEC) || go install github.com/securego/gosec/v2/cmd/gosec@latest
	$(GOSEC) -exclude-generated -severity medium -confidence medium -quiet ./...

docker-build:
	docker build -f $(DOCKERFILE) -t $(IMAGE) .

docker-up:
	$(COMPOSE) up --build -d

docker-test:
	$(COMPOSE) up --build -d --wait
	curl -fsS http://127.0.0.1:8080/healthz
	curl -fsS -o /dev/null http://127.0.0.1:8080/en/
	curl -fsS -o /dev/null http://127.0.0.1:8080/en/search-index.json

docker-down:
	$(COMPOSE) down

ci: tidy
	@test -z "$$(gofmt -l $$(go list -f '{{.Dir}}' ./...))" || (echo "gofmt required" && exit 1)
	go fix ./...
	go vet ./...
	$(MAKE) gosec
	go test ./...
	$(MAKE) build
