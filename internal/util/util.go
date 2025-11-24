package util

import (
	"context"
	"fmt"
	"log"

	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/goldmane"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/otlp"
)

func StartLogStreaming(context context.Context, client *goldmane.GoldmaneClient, logger *otlp.Logger) error {
	data, err := client.StreamFlows(context)
	if err != nil {
		return fmt.Errorf("error during start of log streaming: %s", err)
	}

	logger.ReceiveFlows(context, data)
	return nil
}

func Cleanup(context context.Context, client *goldmane.GoldmaneClient, logger *otlp.Logger) {
	if err := client.Close(); err != nil {
		log.Printf("Error while closing client: %s", err)
	}
	if err := logger.Close(context); err != nil {
		log.Printf("Error while shutting down loggerProvider: %s", err)
	}
}
