GOLDMANE_PROTO_VERSION=v3.30.4
PROTOBUF_DEFINITIONS_DIR=proto
GOLDMANE_CERTIFICATES_DIR=certs/goldmane
GOLDMANE_HOST=goldmane:7443
GOLDMANE_NAMESPACE=calico-system
# Disable HTTPS when running locally
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
GO_PROGRAM=calico-flow-logs-otlphttp-exporter
OUT_DIR=out
CONTAINER_WOKRDIR=/build
TAG=0.0.1-development.1

install-development-packages:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

fetch-protobuf-definition:
	mkdir -p $(PROTOBUF_DEFINITIONS_DIR)
	curl -sL https://raw.githubusercontent.com/projectcalico/calico/refs/tags/$(GOLDMANE_PROTO_VERSION)/goldmane/proto/api.proto -o $(PROTOBUF_DEFINITIONS_DIR)/api.proto

generate-code-from-protobuf:
	which protoc > /dev/null || (echo "protoc not in PATH" && exit 1)
	which protoc-gen-go > /dev/null || (echo "protoc-gen-go not in PATH" && exit 1)
	which protoc-gen-go-grpc > /dev/null || (echo "protoc-gen-go-grpc not in PATH" && exit 1)
	protoc --go_opt=paths=source_relative --go_out=. --go-grpc_out=. $(PROTOBUF_DEFINITIONS_DIR)/*.proto

copy-goldmane-certs-from-kubernetes-deployment:
	mkdir -p $(GOLDMANE_CERTIFICATES_DIR)
	kubectl get secrets goldmane-key-pair -n $(GOLDMANE_NAMESPACE) -o jsonpath='{.data.tls\.crt}' | base64 -d > $(GOLDMANE_CERTIFICATES_DIR)/goldmane.crt
	kubectl get secrets goldmane-key-pair -n $(GOLDMANE_NAMESPACE) -o jsonpath='{.data.tls\.key}' | base64 -d > $(GOLDMANE_CERTIFICATES_DIR)/goldmane.key
	kubectl get cm goldmane-ca-bundle -o jsonpath='{.data.tigera-ca-bundle\.crt}' > $(GOLDMANE_CERTIFICATES_DIR)/goldmane_ca.crt

port-forward-goldmane:
	kubectl port-forward -n $(GOLDMANE_NAMESPACE) svc/goldmane 7443

docker-compose-up:
	docker compose --project-directory hack/docker up

run:
	CA_CERT_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane_ca.crt \
	PRIVATE_KEY_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane.key \
	PUBLIC_CERT_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane.crt \
	GOLDMANE_HOST=$(GOLDMANE_HOST) \
	OTEL_EXPORTER_OTLP_ENDPOINT=$(OTEL_EXPORTER_OTLP_ENDPOINT) \
	go run cmd/calico-flow-logs-otlphttp-exporter/main.go

build:
	go build -C cmd/$(GO_PROGRAM) -o ../../$(OUT_DIR)/$(GO_PROGRAM)

run-built-binary:
	CA_CERT_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane_ca.crt \
	PRIVATE_KEY_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane.key \
	PUBLIC_CERT_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane.crt \
	GOLDMANE_HOST=$(GOLDMANE_HOST) \
	OTEL_EXPORTER_OTLP_ENDPOINT=$(OTEL_EXPORTER_OTLP_ENDPOINT) \
	./$(OUT_DIR)/$(GO_PROGRAM)

build-container-image:
	docker build -t $(GO_PROGRAM):$(TAG) .

run-container: build-container-image
	docker run \
		--network host \
		--name $(GO_PROGRAM) \
		--rm \
		-v ./$(GOLDMANE_CERTIFICATES_DIR):$(CONTAINER_WOKRDIR)/$(GOLDMANE_CERTIFICATES_DIR) \
		-e CA_CERT_PATH=/$(CONTAINER_WOKRDIR)/$(GOLDMANE_CERTIFICATES_DIR)/goldmane_ca.crt \
		-e PRIVATE_KEY_PATH=/$(CONTAINER_WOKRDIR)/$(GOLDMANE_CERTIFICATES_DIR)/goldmane.key \
		-e PUBLIC_CERT_PATH=/$(CONTAINER_WOKRDIR)/$(GOLDMANE_CERTIFICATES_DIR)/goldmane.crt \
		-e GOLDMANE_HOST=$(GOLDMANE_HOST) \
		-e OTEL_EXPORTER_OTLP_ENDPOINT=$(OTEL_EXPORTER_OTLP_ENDPOINT) \
		-it $(GO_PROGRAM):$(TAG)
