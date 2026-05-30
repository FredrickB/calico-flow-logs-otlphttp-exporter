package otlp

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

// Adapted from: https://opentelemetry.io/docs/languages/go/instrumentation/#direct-to-collector
func NewLoggerProvider(
	context context.Context,
	serviceName string,
	serviceVersion string,
) (*log.LoggerProvider, error) {
	resource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating resource: %s", err)
	}

	exporter, err := otlploghttp.New(context)
	if err != nil {
		return nil, fmt.Errorf("error while creating exporter: %s", err)
	}

	loggerProcessor := log.NewBatchProcessor(exporter)
	loggerProvider := log.NewLoggerProvider(
		log.WithResource(resource),
		log.WithProcessor(loggerProcessor),
	)

	return loggerProvider, nil
}
