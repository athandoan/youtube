SERVICES := gateway-service metadata-service upload-service streaming-service
PROTO_DIR := proto
export PATH := $(shell go env GOPATH)/bin:$(PATH)

.PHONY: all lint gen-proto setup-garage tidy build up down logs clean test test-coverage generate-mocks

all: lint build

# ----------------------------------------------------------------------------
# Testing
# ----------------------------------------------------------------------------
test:
	@echo "Running tests..."
	@for service in $(SERVICES); do \
		echo "Testing $$service..."; \
		(cd $$service && go test -v ./...) || exit 1; \
	done
	@echo "✅ All tests passed."

test-coverage:
	@echo "Running tests with coverage..."
	@for service in $(SERVICES); do \
		echo "Testing $$service with coverage..."; \
		(cd $$service && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out) || exit 1; \
	done
	@echo "✅ Coverage reports generated."

# ----------------------------------------------------------------------------
# Mock Generation
# ----------------------------------------------------------------------------
generate-mocks:
	@echo "Checking for mockgen..."
	@command -v mockgen >/dev/null 2>&1 || { echo "Installing mockgen..."; go install go.uber.org/mock/mockgen@latest; }
	@echo "Generating mocks..."
	@for service in $(SERVICES); do \
		echo "Generating mocks for $$service..."; \
		(cd $$service && go generate ./...) || exit 1; \
	done
	@echo "✅ Mocks generated."

# ----------------------------------------------------------------------------
# Code Quality & Dependencies
# ----------------------------------------------------------------------------
lint:
	@echo "Checking for golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	@echo "Linting services..."
	@for service in $(SERVICES); do \
		if [ -d "$$service" ]; then \
			echo "Linting $$service..."; \
			(cd $$service && golangci-lint run) || exit 1; \
		else \
			echo "⚠️  Directory $$service not found!"; \
		fi \
	done
	@echo "✅ All services passed linting."

tidy:
	@echo "Tidying modules..."
	@for service in $(SERVICES) $(PROTO_DIR); do \
		echo "Tidying $$service..."; \
		(cd $$service && go mod tidy) || exit 1; \
	done

# ----------------------------------------------------------------------------
# Proto Generation
# ----------------------------------------------------------------------------
gen-proto:
	@echo "Checking for protoc..."
	@command -v protoc >/dev/null 2>&1 || { echo "❌ protoc not installed. Please install protobuf-compiler."; exit 1; }
	@echo "Checking for protoc-gen-go..."
	@command -v protoc-gen-go >/dev/null 2>&1 || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Initializing proto module if needed..."
	@[ -f $(PROTO_DIR)/go.mod ] || (cd $(PROTO_DIR) && go mod init github.com/athandoan/youtube/proto)
	@echo "Generating Go code from protos..."
	@find $(PROTO_DIR) -name "*.proto" | while read -r proto_file; do \
		echo "Processing $$proto_file..."; \
		protoc --go_out=. --go_opt=paths=source_relative \
		       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
		       "$$proto_file"; \
	done
	@echo "Tidying proto module..."
	@(cd $(PROTO_DIR) && go get google.golang.org/grpc google.golang.org/protobuf && go mod tidy)
	@echo "✅ Proto generation complete."

# ----------------------------------------------------------------------------
# Build & Run
# ----------------------------------------------------------------------------
build:
	@echo "Building services..."
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		(cd $$service && go build -v .) || exit 1; \
	done
	@echo "✅ All services built successfully."

up:
	docker compose up -d garage
	@echo "Waiting for Garage..."
	@sleep 2
	@./setup-garage.sh
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v --rmi local
	@echo "✅ Cleaned up containers, volumes, and local images."

# ----------------------------------------------------------------------------
# Infrastructure Setup
# ----------------------------------------------------------------------------
setup-garage:
	@./setup-garage.sh
