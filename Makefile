PROTO_VERSION=v3.30.4
PROTOBUF_DEFINITIONS_DIR=proto

install-development-packages:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

fetch-protobuf-definition:
	mkdir -p $(PROTOBUF_DEFINITIONS_DIR)
	curl -sL https://raw.githubusercontent.com/projectcalico/calico/refs/tags/$(PROTO_VERSION)/goldmane/proto/api.proto -o $(PROTOBUF_DEFINITIONS_DIR)/api.proto

generate-code-from-protobuf:
	which protoc-gen-go > /dev/null || (echo "protoc-gen-go not in PATH" && exit 1)
	which protoc-gen-go-grpc > /dev/null || (echo "protoc-gen-go-grpc not in PATH" && exit 1)
	protoc --go_opt=paths=source_relative --go_out=. --go-grpc_out=. $(PROTOBUF_DEFINITIONS_DIR)/*.proto

run:
	go run cmd/calico-flow-logs-loki-exporter/main.go
