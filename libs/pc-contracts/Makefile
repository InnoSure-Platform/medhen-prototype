.PHONY: generate clean init

# Install protoc plugins if not present
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate stubs
generate: init
	mkdir -p gen/go
	protoc --proto_path=proto \
		--go_out=gen/go \
		--go-grpc_out=gen/go \
		proto/v1/product.proto \
		proto/v1/party.proto \
		proto/v1/iam.proto \
		proto/v1/integration.proto \
		proto/v1/reporting.proto

clean:
	rm -rf gen/
