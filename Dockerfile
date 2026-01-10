FROM golang:1.25-alpine AS build

ARG VERSION="REPLACE_ME"

RUN apk add --no-cache make protoc curl
WORKDIR /build
COPY . .
RUN make install-development-packages generate-code-from-protobuf lint test build VERSION=${VERSION}

FROM alpine:3.22 AS app

WORKDIR /app
RUN adduser -S exporter
COPY --from=build /build/out/calico-flow-logs-otlphttp-exporter /app/calico-flow-logs-otlphttp-exporter
RUN chown -R exporter /app
USER exporter
ENTRYPOINT ["sh", "-c", "/app/calico-flow-logs-otlphttp-exporter"]
