# Default goal
.DEFAULT_GOAL := help

# Project related variables
PROJECT_NAME := ai-api-proxy
VERSION ?= v0.0.1
DOCKER_USERNAME ?= wangwei01

BASE_PATH := $(shell pwd)
BUILD_PATH := $(BASE_PATH)/build
BINARY_NAME := $(PROJECT_NAME)
MAIN := $(BASE_PATH)/cmd/proxy/main.go

# Environment variables
GO := go
GOCMD := $(GO)
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOFMT := $(GOCMD) fmt
GOLINT := golangci-lint
GO_MOD_TIDY := $(GOCMD) mod tidy

# Build parameters
BUILD_FLAGS := CGO_ENABLED=0 GOOS=linux GOARCH=amd64
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION)"

# Dummy target
.PHONY: all build clean fmt test lint tidy run docker docker-push help

# Default goal
all: build

# Compile binary file
build: clean tidy
	@mkdir -p $(BUILD_PATH)
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@$(BUILD_FLAGS) $(GOBUILD) $(LDFLAGS) -o $(BUILD_PATH)/$(BINARY_NAME) $(MAIN)
	@echo "Build done: $(BUILD_PATH)/$(BINARY_NAME)"

# Clean build files
clean:
	@echo "Clean build files..."
	@rm -rf $(BUILD_PATH)/*
	@$(GOCLEAN)
	@rm -f ./vendor/ ./cover.*

# Format code
fmt:
	@echo "Format code..."
	@$(GOFMT) -s -w ./...

# Run tests
test:
	@echo "Run tests..."
	@$(GOTEST) -v -race -coverprofile=./cover.text -covermode=atomic ./...

# Run static code check
lint:
	@echo "Run code check..."
	@$(GOLINT) run ./...

# Update dependencies
tidy:
	@echo "Update dependencies..."
	@$(GO_MOD_TIDY)

# Run program
run: build
	@echo "Run program..."
	@$(BUILD_PATH)/$(BINARY_NAME)

# Build Docker image
docker:
	@echo "Build Docker image..."
	@docker build -t $(PROJECT_NAME):$(VERSION) .
	@echo "Add tag to docker image..."
	@docker tag $(PROJECT_NAME):$(VERSION) $(DOCKER_USERNAME)/$(PROJECT_NAME):$(VERSION)
	@docker tag $(PROJECT_NAME):$(VERSION) $(DOCKER_USERNAME)/$(PROJECT_NAME):latest
	
# Push Docker image to registry
docker-push: docker
	@echo "Push Docker image to registry..."
	@docker push $(DOCKER_USERNAME)/$(PROJECT_NAME):$(VERSION)
	@docker push $(DOCKER_USERNAME)/$(PROJECT_NAME):latest

# Display help information
help:
	@echo "Available make commands:"
	@echo "  all         - Default goal, build project"
	@echo "  build       - Compile binary file"
	@echo "  clean       - Clean build files"
	@echo "  fmt         - Format code"
	@echo "  test        - Run tests"
	@echo "  lint        - Run code check"
	@echo "  tidy        - Update dependencies"
	@echo "  run         - Run program"
	@echo "  docker      - Build Docker image"
	@echo "  docker-push - Push Docker image to registry"
	@echo "  docker-run  - Run Docker container with external configuration file"
	@echo "  help        - Display help information"


	
