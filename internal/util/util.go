package util

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/goldmane"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/otlp"
	"google.golang.org/grpc"
)

func StartLogStreaming(context context.Context, client goldmane.GoldmaneApi, logger otlp.OtlpLogger, reconnectWaitTime time.Duration) (chan bool, error) {
	done := make(chan bool)

	data, err := client.StreamFlows(context, done, reconnectWaitTime)
	if err != nil {
		return nil, fmt.Errorf("error during start of log streaming: %s", err)
	}

	go logger.ReceiveFlows(context, data)
	return done, nil
}

func Cleanup(context context.Context, client *goldmane.GoldmaneClient, logger *otlp.Logger, connection *grpc.ClientConn) {
	if err := client.Close(); err != nil {
		log.Printf("Error while closing client: %s", err)
	}
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
