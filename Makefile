.PHONY: help build run test clean up start down

help:
	@echo "Distributed Louvain Algorithm - Make targets:"
	@echo "  build        - Build the standalone binary"
	@echo "  run          - Run in standalone mode"
	@echo "  test         - Run tests"
	@echo "  up           - Deploy the distributed-louvain system"
	@echo "  start        - Start the algorithm on the coordinator"
	@echo "  down         - Stop and remove the distributed-louvain system"
	@echo "  clean        - Clean build artifacts"

build:
	@echo "Building standalone binary..."
	go build -o bin/standalone cmd/standalone/main.go

run: build
	@echo "Running standalone mode..."
	./bin/standalone

up:
	@echo "Deploying distributed-louvain..."
	cd deploy && docker compose up --build -d

start:
	@echo "Starting the algorithm on the coordinator..."
	cd deploy && ./start_algorithm.sh localhost:8080 coordinator

down:
	@echo "Stopping distributed-louvain..."
	cd deploy && docker compose down

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf output/
	rm -rf pkg/test/
	go clean

tidy:
	@echo "Tidying modules..."
	go mod tidy
