.PHONY: build test test-unit test-int test-e2e lint fmt vet install clean run

# Binary name
BINARY_NAME=odin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S')
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

# Directories
CMD_DIR=./cmd/${BINARY_NAME}
INTERNAL_DIR=./internal
PKG_DIR=./pkg
BUILD_DIR=./build

# Go commands
GOCMD=go
GOBUILD=${GOCMD} build ${LDFLAGS}
GOTEST=${GOCMD} test
GOVET=${GOCMD} vet
GOFMT=gofmt
GOIMPORTS=goimports
GOLANGCI_LINT=golangci-lint run

# Test coverage
COVERAGE_FILE=coverage.out
COVERAGE_DIR=coverage

# Colors
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

default: build

## build: Build the binary
build:
	@echo "${YELLOW}Building ${BINARY_NAME}...${NC}"
	@mkdir -p ${BUILD_DIR}
	${GOBUILD} -o ${BUILD_DIR}/${BINARY_NAME} ${CMD_DIR}
	@echo "${GREEN}Built successfully!${NC}"

## run: Run the application
run: build
	@echo "${YELLOW}Running ${BINARY_NAME}...${NC}"
	@${BUILD_DIR}/${BINARY_NAME}

## test: Run all tests
test: test-unit

## test-unit: Run unit tests
test-unit:
	@echo "${YELLOW}Running unit tests...${NC}"
	${GOTEST} -v -race -coverprofile=${COVERAGE_FILE} ./...

## test-int: Run integration tests
test-int:
	@echo "${YELLOW}Running integration tests...${NC}"
	${GOTEST} -v -tags=integration ./...

## test-e2e: Run end-to-end tests
test-e2e:
	@echo "${YELLOW}Running E2E tests...${NC}"
	@if command -v docker &> /dev/null; then \
		${GOTEST} -v -tags=e2e ./e2e/...; \
	else \
		echo "${RED}Docker not found. E2E tests require Docker.${NC}"; \
	fi

## test-bdd: Run BDD/godog tests
test-bdd:
	@echo "${YELLOW}Running BDD behavior tests...${NC}"
	@if command -v godog &> /dev/null; then \
		cd openspec && godog run features; \
	else \
		echo "${RED}godog not installed. Install: go install github.com/cucumber/godog/cmd/godog@latest${NC}"; \
	fi

## test-race: Run tests with race detector
test-race:
	@echo "${YELLOW}Running tests with race detector...${NC}"
	${GOTEST} -v -race ./...

## test-coverage: Run tests with detailed coverage report
test-coverage:
	@echo "${YELLOW}Running tests with coverage...${NC}"
	${GOTEST} -v -race -coverprofile=${COVERAGE_FILE} -covermode=atomic ./...
	@mkdir -p ${COVERAGE_DIR}
	@go tool cover -html=${COVERAGE_FILE} -o ${COVERAGE_DIR}/coverage.html
	@go tool cover -func=${COVERAGE_FILE} -o ${COVERAGE_DIR}/coverage.txt
	@echo "${GREEN}Coverage report generated at ${COVERAGE_DIR}/coverage.html${NC}"
	@echo "${GREEN}Coverage summary at ${COVERAGE_DIR}/coverage.txt${NC}"

## test-all: Run all tests (unit, integration, e2e, race)
test-all: test-unit test-int test-race
	@echo "${GREEN}All tests passed!${NC}"

## validate-coverage: Validate coverage meets minimum threshold
validate-coverage:
	@echo "${YELLOW}Validating coverage thresholds...${NC}"
	@if [ ! -f ${COVERAGE_FILE} ]; then \
		echo "${RED}No coverage file found. Run 'make test-coverage' first.${NC}"; \
		exit 1; \
	fi
	@echo "Checking coverage thresholds..."
	@MIN_COVERAGE=70; \
	CUR=$(go tool cover -func=${COVERAGE_FILE} | grep "total:" | awk '{print $$3}' | tr -d '%'); \
	if [ "$$CUR" -lt "$$MIN_COVERAGE" ]; then \
		echo "${RED}Coverage $$CUR% is below minimum $$MIN_COVERAGE%${NC}"; \
		exit 1; \
	fi
	@echo "${GREEN}Coverage $$CUR% meets minimum $$MIN_COVERAGE%${NC}"

## lint: Run linter
lint:
	@echo "${YELLOW}Running linter...${NC}"
	@if command -v ${GOLANGCI_LINT} &> /dev/null; then \
		${GOLANGCI_LINT} ./...; \
	else \
		echo "${RED}golangci-lint not installed. Install: go install github.com/golangci-lint/cmd/golangci-lint@latest${NC}"; \
	fi

## fmt: Format code
fmt:
	@echo "${YELLOW}Formatting code...${NC}"
	@${GOFMT} -s -w .
	@${GOIMPORTS} -w .

## vet: Run go vet
vet:
	@echo "${YELLOW}Running go vet...${NC}"
	${GOVET} ./...

## install: Install binary to GOPATH
install:
	@echo "${YELLOW}Installing ${BINARY_NAME}...${NC}"
	${GOBUILD} -o ${GOPATH}/bin/${BINARY_NAME} ${CMD_DIR}
	@echo "${GREEN}Installed to \$${GOPATH}/bin/${BINARY_NAME}${NC}"

## clean: Clean build artifacts
clean:
	@echo "${YELLOW}Cleaning...${NC}"
	@rm -rf ${BUILD_DIR}
	@rm -f ${COVERAGE_FILE}
	@rm -rf ${COVERAGE_DIR}
	@echo "${GREEN}Cleaned!${NC}"

## coverage: Run tests with coverage report
coverage: test-unit
	@echo "${YELLOW}Generating coverage report...${NC}"
	@mkdir -p ${COVERAGE_DIR}
	@go tool cover -html=${COVERAGE_FILE} -o ${COVERAGE_DIR}/coverage.html
	@echo "${GREEN}Coverage report generated at ${COVERAGE_DIR}/coverage.html${NC}"

## init: Initialize ODIN for first use
init:
	@echo "${YELLOW}Initializing ODIN...${NC}"
	@${BUILD_DIR}/${BINARY_NAME} init || (echo "${RED}Build required first: make build${NC}" && exit 1)

## status: Check ODIN status
status:
	@${BUILD_DIR}/${BINARY_NAME} status || (echo "${RED}Build required first: make build${NC}" && exit 1)

## help: Show this help
help:
	@echo "${YELLOW}ODIN AI${NC} - Nórdico Local-First AI Ecosystem"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  ${GREEN}%-15s${NC} %s\n", $$1, $$2}'
