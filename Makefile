GOLDMANE_PROTO_VERSION=v3.30.4
PROTOBUF_DEFINITIONS_DIR=proto
GEN_DIR=gen
GOLDMANE_CERTIFICATES_DIR=certs/goldmane
GOLDMANE_HOST=goldmane:7443
GOLDMANE_NAMESPACE=calico-system
# Disable HTTPS when running locally
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
GO_PROGRAM=calico-flow-logs-otlphttp-exporter
VERSION=DEVELOPMENT_BUILD
OUT_DIR=out
CONTAINER_WOKRDIR=/build
TAG=0.0.1-development.1
HELM_CHART_REGISTRY_URL=https://raw.githubusercontent.com/FredrickB/calico-flow-logs-otlphttp-exporter/gh-pages
K3D_CLUSTER_NAME=calico-flow-logs
K3D_CLUSTER_CALICO_VERSION=v3.30.4

.PHONY : \
clean \
install-development-packages \
fetch-protobuf-definition \
generate-code-from-protobuf \
build \
copy-goldmane-certs-from-kubernetes-deployment \
port-forward-goldmane \
docker-compose-up \
run \
lint \
test \
debug \
run-built-binary \
run-container \
install-helm-chart-from-local-dir \
setup-private-helm-chart-repository \
install-helm-chart-from-private-chart-repository \
setup-k3d \
install-calico \
start-k3d \
stop-k3d

all: clean install-development-packages fetch-protobuf-definition generate-code-from-protobuf lint test build
.PHONY : all

clean:
	rm -rf out gen

install-development-packages:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

fetch-protobuf-definition:
	mkdir -p $(PROTOBUF_DEFINITIONS_DIR)
	curl -sL https://raw.githubusercontent.com/projectcalico/calico/refs/tags/$(GOLDMANE_PROTO_VERSION)/goldmane/proto/api.proto -o $(PROTOBUF_DEFINITIONS_DIR)/api.proto

generate-code-from-protobuf:
	which protoc > /dev/null || (echo "protoc not in PATH" && exit 1)
	which protoc-gen-go > /dev/null || (echo "protoc-gen-go not in PATH" && exit 1)
	which protoc-gen-go-grpc > /dev/null || (echo "protoc-gen-go-grpc not in PATH" && exit 1)
	mkdir -p $(GEN_DIR)
	protoc \
		--go_out=$(GEN_DIR) \
		--go-grpc_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		$(PROTOBUF_DEFINITIONS_DIR)/api.proto

copy-goldmane-certs-from-kubernetes-deployment:
	mkdir -p $(GOLDMANE_CERTIFICATES_DIR)
	kubectl get secrets goldmane-key-pair -n $(GOLDMANE_NAMESPACE) -o jsonpath='{.data.tls\.crt}' | base64 -d > $(GOLDMANE_CERTIFICATES_DIR)/goldmane.crt
	kubectl get secrets goldmane-key-pair -n $(GOLDMANE_NAMESPACE) -o jsonpath='{.data.tls\.key}' | base64 -d > $(GOLDMANE_CERTIFICATES_DIR)/goldmane.key
	kubectl get cm goldmane-ca-bundle -n $(GOLDMANE_NAMESPACE) -o jsonpath='{.data.tigera-ca-bundle\.crt}' > $(GOLDMANE_CERTIFICATES_DIR)/goldmane_ca.crt

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
	go run cmd/$(GO_PROGRAM)/main.go

debug:
	CA_CERT_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane_ca.crt \
	PRIVATE_KEY_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane.key \
	PUBLIC_CERT_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane.crt \
	GOLDMANE_HOST=$(GOLDMANE_HOST) \
	OTEL_EXPORTER_OTLP_ENDPOINT=$(OTEL_EXPORTER_OTLP_ENDPOINT) \
	dlv debug cmd/$(GO_PROGRAM)/main.go

lint:
	gofmt -l .

test: generate-code-from-protobuf
	go test -v ./...

build:
	go build \
		-C cmd/$(GO_PROGRAM) \
		-o ../../$(OUT_DIR)/$(GO_PROGRAM) \
		--ldflags "\
			-X github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/version.version=$(VERSION) \
			-X github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/version.goldmaneProtobufVersion=$(GOLDMANE_PROTO_VERSION) \
		"

run-built-binary:
	CA_CERT_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane_ca.crt \
	PRIVATE_KEY_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane.key \
	PUBLIC_CERT_PATH=$(GOLDMANE_CERTIFICATES_DIR)/goldmane.crt \
	GOLDMANE_HOST=$(GOLDMANE_HOST) \
	OTEL_EXPORTER_OTLP_ENDPOINT=$(OTEL_EXPORTER_OTLP_ENDPOINT) \
	./$(OUT_DIR)/$(GO_PROGRAM)

build-container-image:
	docker build --build-arg VERSION=$(TAG) -t $(GO_PROGRAM):$(TAG) .

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

install-helm-chart-from-local-dir:
	helm upgrade \
		--install \
		-n $(GOLDMANE_NAMESPACE) \
		-f hack/charts/$(GO_PROGRAM)/override.yaml \
		$(GO_PROGRAM) ./charts/$(GO_PROGRAM)

setup-private-helm-chart-repository:
	if [ -z $${ACCESS_TOKEN:+x} ]; then (echo "Set ACCESS_TOKEN to Personal Access token with read access to repo contents" && exit 1); fi
	helm repo add \
		--username ghp \
		--password $$ACCESS_TOKEN \
		$(GO_PROGRAM) \
		$(HELM_CHART_REGISTRY_URL)
	helm repo update

install-helm-chart-from-private-chart-repository:
	helm upgrade \
		--install \
		-n $(GOLDMANE_NAMESPACE) \
		-f hack/charts/$(GO_PROGRAM)/override.yaml \
		$(GO_PROGRAM) $(GO_PROGRAM)/$(GO_PROGRAM)

setup-k3d:
	k3d cluster create $(K3D_CLUSTER_NAME)-$(K3D_CLUSTER_CALICO_VERSION) \
		--k3s-arg '--flannel-backend=none@server:*' \
		--k3s-arg '--disable-network-policy@server:*' \
		--k3s-arg '--cluster-cidr=192.168.0.0/16@server:*'
	kubectl apply -R -f hack/k3d

install-calico:
	kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/$(K3D_CLUSTER_CALICO_VERSION)/manifests/tigera-operator.yaml
	kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/$(K3D_CLUSTER_CALICO_VERSION)/manifests/custom-resources.yaml

start-k3d:
	k3d cluster start $(K3D_CLUSTER_NAME)-$(K3D_CLUSTER_CALICO_VERSION)

stop-k3d:
	k3d cluster stop $(K3D_CLUSTER_NAME)-$(K3D_CLUSTER_CALICO_VERSION)
