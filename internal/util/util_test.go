package util

import (
	"context"
	"sync"
	"testing"
	"time"

	mocks "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/mocks"
	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/protos"

	"go.uber.org/mock/gomock"
)

var (
	flow = pb.Flow{
		SourceLabels: []string{"test=true"},
	}
)

func TestShouldParseSecondsCorrectly(t *testing.T) {
	secondStringValue := "5"
	defaultValue := 10 * time.Second
	expected := 5 * time.Second
	actual := ParseSecondsStringValue(secondStringValue, defaultValue)

	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}

func TestInvalidSecondsReturnsDefaultValue(t *testing.T) {
	secondStringValue := "invalid seconds"
	defaultValue := 10 * time.Second
	expected := defaultValue
	actual := ParseSecondsStringValue(secondStringValue, defaultValue)

	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}

func TestStartLogStreaming(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	dataChMock := make(chan *pb.Flow, 1)
	dataChMock <- &flow

	goldmaneApiMock := mocks.NewMockGoldmaneApi(ctrl)
	goldmaneApiMock.EXPECT().
		StreamFlows(ctx, gomock.Any()).
		Return(dataChMock, nil).
		Times(1)

	otlpLoggerMock := mocks.NewMockOtlpLogger(ctrl)
	otlpLoggerMock.EXPECT().
		ReceiveFlows(ctx, dataChMock).
		// replace method with a call to the waitgroup
		// since the real implementation is ran in its
		// own goroutine
		Do(func(_ context.Context, _ <-chan *pb.Flow) {
			wg.Done()
		}).
		Times(1)

	reconnects := StartLogStreaming(ctx, goldmaneApiMock, otlpLoggerMock, 2*time.Second)

	if reconnects != nil {
		t.Errorf("reconnects should be nil")
	}

	// block and wait for the otlplogger to be invoked
	// in its own goroutine
	wg.Wait()
}
