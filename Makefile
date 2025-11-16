GOLDMANE_PROTO_VERSION=v3.30.4
PROTOBUF_DEFINITIONS_DIR=proto
GOLDMANE_CERTIFICATES_DIR=certs/goldmane
GOLDMANE_HOST=goldmane:7443
GOLDMANE_NAMESPACE=calico-system
# Disable HTTPS when running locally
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

install-development-packages:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

fetch-protobuf-definition:
	mkdir -p $(PROTOBUF_DEFINITIONS_DIR)
	curl -sL https://raw.githubusercontent.com/projectcalico/calico/refs/tags/$(GOLDMANE_PROTO_VERSION)/goldmane/proto/api.proto -o $(PROTOBUF_DEFINITIONS_DIR)/api.proto

generate-code-from-protobuf:
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

run-otel-collector:
	docker compose --project-directory hack/otel-collector up

run:
	CA_CERT_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane_ca.crt \
	PRIVATE_KEY_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane.key \
	PUBLIC_CERT_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane.crt \
	GOLDMANE_HOST=$(GOLDMANE_HOST) \
	OTEL_EXPORTER_OTLP_ENDPOINT=$(OTEL_EXPORTER_OTLP_ENDPOINT) \
	go run cmd/calico-flow-logs-loki-exporter/main.go
