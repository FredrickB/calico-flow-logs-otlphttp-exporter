package util

import (
	"context"
	"fmt"
	"log"

	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/goldmane"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/otlp"
)

func StartLogStreaming(context context.Context, client *goldmane.GoldmaneClient, logger *otlp.Logger) (chan bool, error) {
	done := make(chan bool)

	data, err := client.StreamFlows(context, done)
	if err != nil {
		return nil, fmt.Errorf("error during start of log streaming: %s", err)
	}

	logger.ReceiveFlows(context, data)
	return done, nil
}

func Cleanup(context context.Context, client *goldmane.GoldmaneClient, logger *otlp.Logger) {
	if err := client.Close(); err != nil {
		log.Printf("Error while closing client: %s", err)
	}
	if err := logger.Close(context); err != nil {
		log.Printf("Error while shutting down loggerProvider: %s", err)
	}
}
