.PHONY: help build run test clean

help:
	@echo "Distributed Louvain Algorithm - Make targets:"
	@echo "  build        - Build the standalone binary"
	@echo "  run          - Run in standalone mode"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"

build:
	@echo "Building standalone binary..."
	go build -o bin/standalone cmd/standalone/main.go

run: build
	@echo "Running standalone mode..."
	./bin/standalone

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf output/
	go clean

tidy:
	@echo "Tidying modules..."
	go mod tidy
