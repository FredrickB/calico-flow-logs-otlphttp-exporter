# calico-flow-logs-otlphttp-exporter

**Currently in development, things will break, run at your own risk**

Stream network flow logs from [Calico](https://docs.tigera.io/calico/latest/observability/flow-logs-api)
using [OTLP](https://opentelemetry.io/docs/specs/otlp/) over HTTP
to [OTel collector](https://opentelemetry.io/docs/collector/).
Uses the [Direct to Collector](https://opentelemetry.io/docs/specs/otel/logs/#direct-to-collector)
approach with the [Logs SDK](https://opentelemetry.io/docs/specs/otel/logs/sdk/)
and [Logging bridge](https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog).

Compatibility:

|Calico version|Tested|
|:---|:---|
|`3.30`|yes|
|`3.31`|no|

## Development

### Prerequisites

- `make`
- `protoc`, `3.21.12`
- `kubectl`
- `base64`
- `docker`
- Calico
- Goldmane running in namespace `calico-system`

### Setup

1. Install development tools: `make install-development-packages`
1. Fetch the protobufs from the Calico project: `make fetch-protobuf-definition`
1. Generate code from protobufs `make generate-code-from-protobuf`
1. Add line to `/etc/hosts` in order to be able to use Goldmane certs for running locally:
    ```
    127.0.0.1 goldmane
    ```

### Running

> For development we copy the certificates from a Goldmane
deployment running in the cluster directly. Do not use this
approach in production environments, this is just for development.

1. Port-forward the Goldmane service: `make [GOLDMANE_NAMESPACE=<goldmane namespace>] port-forward-goldmane`
1. Copy the certificates from a running instance of Goldmane: `make [GOLDMANE_NAMESPACE=<goldmane namespace>] copy-goldmane-certs-from-kubernetes-deployment`
1. Start the otel-collector, Loki and Grafana: `make docker-compose-up`
1. Run project: `make run`
1. [Open Grafana Explore with Loki search + JSON parsing enabled](http://localhost:3000/explore?schemaVersion=1&panes=%7B%22fns%22:%7B%22datasource%22:%22P8E80F9AEF21F6940%22,%22queries%22:%5B%7B%22refId%22:%22A%22,%22expr%22:%22%7Bservice_name%3D%5C%22calico-flow-logs-otlphttp-exporter%5C%22%7D%20%7C%20json%22,%22queryType%22:%22range%22,%22datasource%22:%7B%22type%22:%22loki%22,%22uid%22:%22P8E80F9AEF21F6940%22%7D,%22editorMode%22:%22code%22,%22direction%22:%22backward%22%7D%5D,%22range%22:%7B%22from%22:%22now-1h%22,%22to%22:%22now%22%7D,%22panelsState%22:%7B%22logs%22:%7B%22visualisationType%22:%22logs%22%7D%7D,%22compact%22:false%7D%7D&orgId=1)
    - Username: `admin123`
    - Password: `admin123`

#### Running as container

1. Port-forward the Goldmane service: `make [GOLDMANE_NAMESPACE=<goldmane namespace>] port-forward-goldmane`
1. Copy the certificates from a running instance of Goldmane: `make [GOLDMANE_NAMESPACE=<goldmane namespace>] copy-goldmane-certs-from-kubernetes-deployment`
1. Start the otel-collector, Loki and Grafana: `make docker-compose-up`
1. Build container image: `make build-container-image`
1. Run the container image: `make run-container`

#### Running as Helm release

##### Prerequisites

1. OpenTelemetry-Collector installed:
   1. Namespace must be `opentelemetry-collector`
   1. Name of Service for OpenTelemetry-Collector must be `opentelemetry-collector`
1. Login to ghcr.io:
    ```bash
    docker login ghcr.io -u ghp
    <Paste Personal Access Token>
    ```
1. Create an imagepullsecret from docker config:
    ```bash
    kubectl create secret generic ghcr-io-regcred \
    -n calico-system \
    --from-file=.dockerconfigjson=$HOME/.docker/config.json \
    --type kubernetes.io/dockerconfigjson
    ```
1. (Optional) Adapt values in `hack/charts/calico-flow-logs-otlphttp-exporter/override.yaml`

##### Install Helm chart from local directory

1. Install Helm release:
    ```bash
    make install-helm-chart-from-local-dir
    ```

##### Install Helm chart from private registry

1. Login to GitHub Packages private chart repository
    ```bash
    read -s ACCESS_TOKEN
    <Paste Personal Access Token with read-only access for repository content>
    ```
1. Setup private helm chart repository:
    ```bash
    make setup-private-helm-chart-repository
    ```
1. Install Helm release:
    ```bash
    make install-helm-chart-from-private-chart-repository
    ```

### OTLP Log HTTP Exporter environment variables

To see a list of all environment variables which can be set for the OTLP
Log HTTP exporter, see the [official documentation](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp)
