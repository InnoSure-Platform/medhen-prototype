.PHONY: proto build run test

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       api/proto/v1/notification.proto

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/server cmd/server/main.go

test:
	go test -v ./...

run:
	go run cmd/server/main.go
