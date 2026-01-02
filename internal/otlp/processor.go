package otlp

import (
	"context"
	"log/slog"

	otelslog "go.opentelemetry.io/contrib/bridges/otelslog"
	otelloggersdk "go.opentelemetry.io/otel/sdk/log"
)

type Processor interface {
	Log(ctx context.Context, message string)
	Close(ctx context.Context) error
}

type OtlpProcessor struct {
	logger         *slog.Logger
	loggerprovider *otelloggersdk.LoggerProvider
}

func NewProcessor(context context.Context, packageName string, loggerProvider *otelloggersdk.LoggerProvider) *OtlpProcessor {
	return &OtlpProcessor{
		logger:         otelslog.NewLogger(packageName, otelslog.WithLoggerProvider(loggerProvider)),
		loggerprovider: loggerProvider,
	}
}

func (p *OtlpProcessor) Log(ctx context.Context, message string) {
	p.logger.Log(ctx, slog.LevelInfo, message)
}

func (p *OtlpProcessor) Close(ctx context.Context) error {
	return p.loggerprovider.Shutdown(ctx)
}
