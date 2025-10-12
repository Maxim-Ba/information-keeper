.PHONY: lint lint-fix vet fmt check help

# Определяем ОС
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
else
    DETECTED_OS := $(shell uname -s)
endif

lint:
	@echo "Running lint on $(DETECTED_OS)..."
	golangci-lint run

lint-fix:
	golangci-lint run --fix

vet:
	go vet ./...

fmt:
	gofmt -w .

check: lint vet

help:
	@echo "Available commands:"
	@echo "  lint      - Run golangci-lint"
	@echo "  lint-fix  - Run golangci-lint with auto-fix"
	@echo "  vet       - Run go vet"
	@echo "  fmt       - Format code with gofmt"
	@echo "  check     - Run all checks"
