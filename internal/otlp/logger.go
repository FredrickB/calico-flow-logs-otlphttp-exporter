package otlp

import (
	"context"
	"errors"
	"log"

	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/proto"

	"google.golang.org/protobuf/encoding/protojson"
)

var (
	ErrContextCancelled  = errors.New("Cancellation invoked, stopping log forwarding")
	ErrDataChannelClosed = errors.New("Data channel closed, stopping log forwarding")
)

type OtlpLogger interface {
	ReceiveFlows(context context.Context, data <-chan *pb.Flow) error
}

type Logger struct {
	processor Processor
}

func NewLogger(processor Processor) *Logger {
	return &Logger{processor: processor}
}

func (logger *Logger) ReceiveFlows(context context.Context, data <-chan *pb.Flow) error {
	for {
		select {
		case <-context.Done():
			return ErrContextCancelled
		case flow, ok := <-data:
			if !ok {
				return ErrDataChannelClosed
			}
			jsonFlow, err := protojson.Marshal(flow)
			if err != nil {
				log.Printf("Failed to marshal flow: %+v to JSON. Error: %s. Skipping log", flow, err)
				continue
			}
			logger.processor.Log(context, string(jsonFlow))
		}
	}
}

func (logger *Logger) Close(context context.Context) error {
	return logger.processor.Close(context)
}
