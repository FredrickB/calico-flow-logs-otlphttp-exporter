package util

import (
	"context"
	"reflect"
	"testing"
	"time"

	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/protos"
)

var (
	flow = pb.Flow{
		SourceLabels: []string{"test=true"},
	}
	flows                   = make(chan *pb.Flow)
	errors                  = make(chan error)
	otlpLoggerReceivedFlows = []*pb.Flow{}
	otlpLoggerError         error
)

type FakeGoldmaneApi struct {
	Flows  chan (*pb.Flow)
	Errors chan (error)
}

func (client *FakeGoldmaneApi) StreamFlows(
	context context.Context,
	reconnectWaitTime time.Duration,
) (<-chan *pb.Flow, <-chan error) {
	return client.Flows, client.Errors
}

type FakeOtlpLogger struct {
	ReceivedFlows []*pb.Flow
}

func (logger *FakeOtlpLogger) ReceiveFlows(context context.Context, data <-chan *pb.Flow) error {
	logger.ReceivedFlows = append(logger.ReceivedFlows, <-data)
	return otlpLoggerError
}

func TestShouldParseSecondsCorrectly(t *testing.T) {
	milliSecondStringValue := "5"
	defaultValue := 10 * time.Millisecond
	expected := 5 * time.Millisecond
	actual := ParseMilliSecondsStringValue(milliSecondStringValue, defaultValue)

	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}

func TestInvalidSecondsReturnsDefaultValue(t *testing.T) {
	milliSecondStringValue := "invalid milliseconds"
	defaultValue := 10 * time.Millisecond
	expected := defaultValue
	actual := ParseMilliSecondsStringValue(milliSecondStringValue, defaultValue)

	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}

func TestStartLogStreaming(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	goldmaneApiFake := &FakeGoldmaneApi{
		Flows:  flows,
		Errors: errors,
	}

	otlpLoggerFake := FakeOtlpLogger{
		ReceivedFlows: otlpLoggerReceivedFlows,
	}

	_ = StartLogStreaming(ctx, goldmaneApiFake, &otlpLoggerFake, 2*time.Second)

	// pass in a flow to trigger consumption in logger
	flows <- &flow

	receivedFlow := *otlpLoggerFake.ReceivedFlows[0]
	if !reflect.DeepEqual(receivedFlow, flow) {
		t.Errorf("expected: %v, actual: %v", flow, receivedFlow)
	}
}
