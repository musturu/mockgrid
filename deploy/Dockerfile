# Multi-stage Dockerfile for mockgrid

# Build stage
FROM golang:1.25.3 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Install CA certs so we can copy the system bundle into the final scratch image.
# Keep this small and clean up apt lists to reduce image size.
RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates \
	&& rm -rf /var/lib/apt/lists/*
# Build a statically linked binary for linux/amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o /mockgrid ./

FROM scratch
USER 10001
COPY --from=builder /mockgrid /mockgrid
# copy CA certificates from the builder so TLS validation works in the final image
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt


ENTRYPOINT ["/mockgrid serve"]
