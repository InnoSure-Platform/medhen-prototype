.PHONY: all build test clean run

APP_NAME=pc-document-mgmt-svc

all: clean test build

build:
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) ./cmd/server

test:
	@echo "Running tests..."
	go test ./... -v

clean:
	@echo "Cleaning up..."
	rm -rf bin/

run: build
	@echo "Running $(APP_NAME)..."
	./bin/$(APP_NAME)
