FROM golang:1.24-alpine AS build

RUN apk add --no-cache make protoc curl
WORKDIR /build
COPY . .
RUN make install-development-packages generate-code-from-protobuf build

FROM alpine:3.22 AS app

COPY --from=build /build/out/calico-flow-logs-otlphttp-exporter /bin/calico-flow-logs-otlphttp-exporter
ENTRYPOINT ["sh", "-c", "/bin/calico-flow-logs-otlphttp-exporter"]
