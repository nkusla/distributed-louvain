.PHONY: help build run test clean deploy

help:
	@echo "Distributed Louvain Algorithm - Make targets:"
	@echo "  build        - Build the standalone binary"
	@echo "  run          - Run in standalone mode"
	@echo "  test         - Run tests"
	@echo "  deploy       - Deploy the distributed-louvain system"
	@echo "  stop         - Stop the distributed-louvain system"
	@echo "  clean        - Clean build artifacts"

build:
	@echo "Building standalone binary..."
	go build -o bin/standalone cmd/standalone/main.go

run: build
	@echo "Running standalone mode..."
	./bin/standalone

deploy:
	@echo "Deploying distributed-louvain..."
	docker compose up --build -d

stop:
	@echo "Stopping distributed-louvain..."
	docker compose down

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
