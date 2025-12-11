# Multi-stage Dockerfile for mockgrid

# Build stage
FROM golang:1.25.5 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Install CA certs and build dependencies for CGO (SQLite support)
RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates gcc libc6-dev sqlite3 busybox \
	&& rm -rf /var/lib/apt/lists/*
# Build with CGO enabled for SQLite support
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o /mockgrid ./

FROM gcr.io/distroless/cc AS runtime
COPY --from=builder /mockgrid /mockgrid
COPY --from=builder /usr/bin/sqlite3 /usr/bin/sqlite3
COPY --from=builder /bin/busybox /bin/busybox
COPY --from=builder /lib/x86_64-linux-gnu /lib/x86_64-linux-gnu
COPY --from=builder /usr/lib/x86_64-linux-gnu /usr/lib/x86_64-linux-gnu
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
RUN ln -sf /bin/busybox /bin/sh
COPY scripts/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["serve"]
