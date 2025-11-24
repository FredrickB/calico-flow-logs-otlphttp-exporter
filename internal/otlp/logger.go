package otlp

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/proto"
	otelslog "go.opentelemetry.io/contrib/bridges/otelslog"
	otelloggersdk "go.opentelemetry.io/otel/sdk/log"

	"google.golang.org/protobuf/encoding/protojson"
)

type Logger struct {
	logger         *slog.Logger
	loggerprovider *otelloggersdk.LoggerProvider
}

func NewLogger(context context.Context, packageName, serviceName, serviceVersion string) (*Logger, error) {
	loggerProvider, err := newLoggerProvider(context, serviceName, serviceVersion)
	if err != nil {
		return nil, fmt.Errorf("error while creating OTLP LoggerProvider %s", err)
	}
	return &Logger{
		logger:         otelslog.NewLogger(packageName, otelslog.WithLoggerProvider(loggerProvider)),
		loggerprovider: loggerProvider,
	}, nil
}

func (l *Logger) ReceiveFlows(context context.Context, receiver <-chan *pb.Flow) {
	go func() {
		for {
			select {
			case <-context.Done():
				log.Printf("Cancellation invoked. Stopping")
				return
			case flow, ok := <-receiver:
				if !ok {
					log.Println("Data channel closed. Stopping")
					return
				}
				jsonFlow, err := protojson.Marshal(flow)
				if err != nil {
					log.Printf("Failed to marshal flow: %+v to JSON. Error: %s. Skipping", flow, err)
					continue
				}
				l.logger.Log(context, slog.LevelInfo, string(jsonFlow))
			}
		}
	}()
}

func (l *Logger) Close(context context.Context) error {
	return l.loggerprovider.Shutdown(context)
}
