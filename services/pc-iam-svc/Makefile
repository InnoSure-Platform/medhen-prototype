.PHONY: build run test

build:
	go build -o bin/pc-iam-svc ./cmd/api

run: build
	./bin/pc-iam-svc

test:
	go test ./... -v
