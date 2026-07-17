# Makefile for pc-claims-svc

.PHONY: build run test clean db-up db-down

APP_NAME=pc-claims-svc
VERSION=1.0.0

build:
	go build -o bin/server cmd/server/main.go

run: build
	./bin/server

test:
	go test -v -cover ./...

clean:
	rm -rf bin/

db-up:
	docker-compose up -d db

db-down:
	docker-compose down

fmt:
	go fmt ./...
