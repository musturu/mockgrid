# Multi-stage Dockerfile for mockgrid

# Build stage
FROM golang:1.25.3 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Install CA certs and build dependencies for CGO (SQLite support)
RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates gcc libc6-dev \
	&& rm -rf /var/lib/apt/lists/*
# Build with CGO enabled for SQLite support
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o /mockgrid ./

# Runtime stage - use minimal image with libc for CGO binary
FROM debian:bookworm-slim
RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates \
	&& rm -rf /var/lib/apt/lists/* \
	&& useradd -u 10001 -m appuser
USER appuser
COPY --from=builder /mockgrid /mockgrid

ENTRYPOINT ["/mockgrid", "serve"]
