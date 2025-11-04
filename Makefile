.PHONY: build run test clean download-sample

# Build the server
build:
	go build -o nav-server cmd/server/main.go

# Run the server
run:
	go run cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f nav-server
	rm -f graph.bin.gz

# Download sample OSM data (Monaco - small dataset)
download-sample:
	@echo "Downloading Monaco OSM data (~1MB)..."
	curl -o monaco-latest.osm.pbf https://download.geofabrik.de/europe/monaco-latest.osm.pbf
	@echo "Download complete!"

# Run with sample data
run-sample: download-sample
	OSM_DATA_PATH=monaco-latest.osm.pbf go run cmd/server/main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Build for production
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nav-server cmd/server/main.go

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the server binary"
	@echo "  run            - Run the server"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  download-sample - Download sample OSM data (Monaco)"
	@echo "  run-sample     - Download and run with sample data"
	@echo "  deps           - Download dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  build-prod     - Build for production"

