# Calico Flow Logs Loki Exporter

Uses the Goldmane Flow logs in Calico to
extract realtime network flow logs and ingest
them into Loki.

## Development

### Prerequisites

- `protoc`, `3.21.12`

### Setup

#### Protobufs

1. Install development tools: `make install-development-packages`
2. Fetch the protobufs from the Calico project: `make fetch-protobuf-definition`
3. Generate code from protobufs `make generate-code-from-protobuf`
