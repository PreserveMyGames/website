# syntax=docker/dockerfile:1

FROM golang:1.26.5-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
  -trimpath \
  -ldflags="-s -w -X github.com/PreserveMyGames/store/internal/web.assetVersion=${VERSION}" \
  -o /out/store \
  ./cmd/store

FROM alpine:3.24@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b AS runtime

RUN apk add --no-cache ca-certificates wget \
  && addgroup -g 65532 -S app \
  && adduser -u 65532 -S app -G app \
  && mkdir -p /data \
  && chown 65532:65532 /data

WORKDIR /app

COPY --from=builder /out/store /app/store

USER 65532:65532

EXPOSE 8080
VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1

LABEL org.opencontainers.image.title="PreserveMyGames Store" \
  org.opencontainers.image.description="Independent store for PreserveMyGames" \
  org.opencontainers.image.source="https://github.com/PreserveMyGames/website" \
  org.opencontainers.image.url="https://store.preservemygames.org" \
  org.opencontainers.image.vendor="PreserveMyGames" \
  org.opencontainers.image.licenses="MIT"

ENV PORT=8080 \
  APP_ENV=production \
  DATA_DIR=/data \
  GODEBUG=multipathtcp=0

ENTRYPOINT ["/app/store"]
