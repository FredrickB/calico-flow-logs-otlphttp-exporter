package util

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/goldmane"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/otlp"
	"google.golang.org/grpc"
)

func StartLogStreaming(
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

func LogEnvironmentVariable(variable, value string) {
	log.Printf("%s set to %s", variable, value)
}

func ParseSecondsStringValue(secondsAsString string, defaultValue time.Duration) time.Duration {
	secondsParsed, err := strconv.Atoi(secondsAsString)
	if err != nil {
		return defaultValue
	} else {
		return time.Duration(secondsParsed) * time.Second
	}
}
