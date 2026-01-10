FROM golang:1.25-alpine AS build

ARG VERSION="REPLACE_ME"

RUN apk add --no-cache make protoc curl
WORKDIR /build
COPY . .
RUN  sed -i "s/\"\(REPLACED_DURING_BUILD\)\"/\"${VERSION}\"/g" cmd/calico-flow-logs-otlphttp-exporter/main.go
RUN make install-development-packages generate-code-from-protobuf lint test build

FROM alpine:3.22 AS app

WORKDIR /app
RUN adduser -S exporter
COPY --from=build /build/out/calico-flow-logs-otlphttp-exporter /app/calico-flow-logs-otlphttp-exporter
RUN chown -R exporter /app
USER exporter
ENTRYPOINT ["sh", "-c", "/app/calico-flow-logs-otlphttp-exporter"]
