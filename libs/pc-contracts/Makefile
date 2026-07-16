.PHONY: generate clean init

# Install protoc plugins if not present
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate stubs
generate: init
	mkdir -p gen/go
	protoc --proto_path=proto \
		--go_out=gen/go --go_opt=paths=source_relative \
		--go-grpc_out=gen/go --go-grpc_opt=paths=source_relative \
		proto/v1/product.proto \
		proto/v1/party.proto

clean:
	rm -rf gen/
