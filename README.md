# Calico Flow Logs Loki Exporter

Uses the Goldmane Flow logs in Calico to
extract realtime network flow logs and ingest
them into Loki.

## Development

### Prerequisites

- `make`
- `protoc`, `3.21.12`
- `kubectl`
- `base64`
- Running instance of Goldmane in namespace `calico-system`

### Setup

**Note: For development we copy the certificates from a Goldmane
deployment running in the cluster directly. Do not use this approach
in production environments, this is just for development.**

1. Install development tools: `make install-development-packages`
1. Fetch the protobufs from the Calico project: `make fetch-protobuf-definition`
1. Generate code from protobufs `make generate-code-from-protobuf`
1. Add line to `/etc/hosts` in order to be able to use Goldmane certs for running locally:
    ```
    127.0.0.1 goldmane
    ```

### Running

1. Port-forward the Goldmane service: `kubectl port-forward [-n <goldmane-namespace>] svc/goldmane 7443`
1. Copy the certificates from a running instance of Goldmane: `make copy-goldmane-certs-from-kubernetes-deployment`
    - The Goldmane deployment must use default naming conventions for resources, otherwise you need to override `GOLDMANE_NAMESPACE`
    when running `make`
1. Run project: `make run`
