package otlp

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
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
		return nil, err
	}

	exporter, err := otlploghttp.New(context)
	if err != nil {
		return nil, err
	}

	loggerProcessor := log.NewBatchProcessor(exporter)
	loggerProvider := log.NewLoggerProvider(
		log.WithResource(resource),
		log.WithProcessor(loggerProcessor),
	)

	return loggerProvider, nil
}
