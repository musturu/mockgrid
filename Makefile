.PHONY: build clean test run docker-build docker-run help

# Variables
BINARY_NAME=mockgrid
VERSION?=latest
DOCKER_IMAGE=mockgrid:$(VERSION)
DOCKER_FILE=deploy/Dockerfile

help:
	@echo "Available targets:"
	@echo "  make build        - Build the mockgrid binary"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make test         - Run all tests"
	@echo "  make run          - Run the mockgrid server"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) ./

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	go clean -cache -testcache

test:
	@echo "Running tests..."
	go test -v -race ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME) serve

docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -f $(DOCKER_FILE) -t $(DOCKER_IMAGE) .
	@echo "Docker image built successfully"

docker-run: docker-build
	@echo "Running Docker container..."
	docker run -it --rm \
		-p 8080:8080 \
		-e SMTP_SERVER=localhost \
		-e SMTP_PORT=1025 \
		$(DOCKER_IMAGE)

.DEFAULT_GOAL := help
