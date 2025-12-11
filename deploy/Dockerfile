# Multi-stage Dockerfile for mockgrid

# Build stage on Alpine so we can reuse musl-linked artifacts
FROM golang:1.25.5-alpine3.21 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Install CA certs and build dependencies for CGO (SQLite support)
RUN apk add --no-cache ca-certificates build-base sqlite-dev sqlite
# Build with CGO enabled for SQLite support
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o /mockgrid ./

# Runtime stage uses lightweight Alpine with busybox/sh
FROM alpine:3.21 AS runtime
RUN apk add --no-cache ca-certificates
COPY --from=builder /mockgrid /mockgrid
COPY --from=builder /usr/bin/sqlite3 /usr/bin/sqlite3
RUN adduser -D -u 10001 appuser
COPY scripts/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
USER appuser

ENTRYPOINT ["/bin/sh", "/docker-entrypoint.sh"]
CMD ["serve"]
