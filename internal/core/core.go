package core

import (
	"context"
	"log"
	"time"

	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/goldmane"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/otlp"
	"google.golang.org/grpc"
)

// Stream flows from Goldmane to logger,
// return channel to receive reconnect
// events.
func Run(
	context context.Context,
	client goldmane.GoldmaneApi,
	logger otlp.OtlpLogger,
	reconnectWaitTime time.Duration,
) <-chan error {
	data, reconnects := client.StreamFlows(context, reconnectWaitTime)
	go logger.ReceiveFlows(context, data)
	return reconnects
}

func Cleanup(context context.Context, logger *otlp.Logger, connection *grpc.ClientConn) {
	if err := connection.Close(); err != nil {
		log.Printf("Error while closing connection: %s", err)
	}
	if err := logger.Close(context); err != nil {
		log.Printf("Error while shutting down loggerProvider: %s", err)
	}
}
