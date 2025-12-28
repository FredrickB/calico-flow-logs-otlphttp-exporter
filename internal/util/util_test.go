package util

import (
	"context"
	"errors"
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
		StreamFlows(ctx, gomock.Any(), gomock.Any()).
		Return(dataChMock, nil, nil).
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

	doneChannel, _, err := StartLogStreaming(ctx, goldmaneApiMock, otlpLoggerMock, 2*time.Second)

	if err != nil {
		t.Errorf("err should be nil")
	}
	if doneChannel == nil {
		t.Errorf("done channel should not be nil")
	}

	// block and wait for the otlplogger to be invoked
	// in its own goroutine
	wg.Wait()
}

func TestStartLogStreamingFailsIfStreamFlowsReturnsErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flowChannelMock := make(<-chan *pb.Flow)
	err := errors.New("mocked error")
	reconnectWaitTime := 2 * time.Second

	goldmaneApiMock := mocks.NewMockGoldmaneApi(ctrl)
	goldmaneApiMock.EXPECT().StreamFlows(ctx, gomock.Any(), gomock.Any()).Return(nil, nil, err)

	otlpLoggerMock := mocks.NewMockOtlpLogger(ctrl)
	otlpLoggerMock.EXPECT().ReceiveFlows(ctx, flowChannelMock).Times(0)

	doneChannel, _, err := StartLogStreaming(ctx, goldmaneApiMock, otlpLoggerMock, reconnectWaitTime)

	if err == nil {
		t.Errorf("err should not be nil")
	}
	if doneChannel != nil {
		t.Errorf("done channel should be nil since error was encountered")
	}
}
