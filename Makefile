.PHONY: server client deps clean proto test docker-build docker-run docker-stop docker-clean compose-up compose-down compose-test compose-logs help

# Default target
help:
	@echo "Available targets:"
	@echo "  deps                          - Download dependencies"
	@echo "  server                        - Run the gRPC server"
	@echo "  client                        - Run the gRPC client"
	@echo "  test                          - Run comprehensive grpcurl tests"
	@echo "  proto                         - Regenerate protocol buffer code"
	@echo "  build                         - Build binaries"
	@echo "  clean                         - Clean build artifacts"
	@echo "  docker-build                  - Build Docker image"
	@echo "  docker-build-multiplatform    - Build and push multiplatform Docker image"
	@echo "  docker-build-multiplatform-local - Build multiplatform image for local use"
	@echo "  docker-run                    - Run Docker container"
	@echo "  docker-stop                   - Stop Docker container"
	@echo "  docker-clean                  - Remove Docker image and container"
	@echo "  compose-up                    - Start services with Docker Compose"
	@echo "  compose-down                  - Stop services with Docker Compose"
	@echo "  compose-test                  - Run tests with Docker Compose"
	@echo "  compose-logs                  - View Docker Compose logs"
	@echo "  compose-multiplatform         - Build and run multiplatform with Docker Compose"
	@echo "  help                          - Show this help message"

# Download dependencies
deps:
	go mod tidy

# Run the server
server:
	go run server/main.go

# Run the client
client:
	go run client/main.go

# Run comprehensive grpcurl tests
test:
	./test.sh

# Regenerate protocol buffer code
proto:
	PATH=$$PATH:~/go/bin protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/hello/hello.proto proto/goodbye/goodbye.proto

# Clean build artifacts
clean:
	go clean
	rm -f server/server client/client

# Build binaries
build:
	go build -o server/server server/main.go
	go build -o client/client client/main.go
