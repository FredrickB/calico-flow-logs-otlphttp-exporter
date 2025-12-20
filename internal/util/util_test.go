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
	flow = pb.Flow{}
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

	dataChannelMock := make(chan *pb.Flow, 1)
	dataChannelMock <- &flow

	goldmaneApiMock := mocks.NewMockGoldmaneApi(ctrl)
	goldmaneApiMock.EXPECT().StreamFlows(ctx, gomock.Any(), gomock.Any()).Return(dataChannelMock, nil).Times(1)

	otlpLoggerMock := mocks.NewMockOtlpLogger(ctrl)
	otlpLoggerMock.EXPECT().ReceiveFlows(ctx, dataChannelMock).Times(1)

	doneChannel, err := StartLogStreaming(ctx, goldmaneApiMock, otlpLoggerMock, 2*time.Second)

	if err != nil {
		t.Errorf("err should be nil")
	}
	if doneChannel == nil {
		t.Errorf("done channel should not be nil")
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// make a waiting mechanism
	go func() {
		doneChannel <- true
		wg.Done()
	}()
	go func() {
		<-doneChannel
		wg.Done()
	}()

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
	goldmaneApiMock.EXPECT().StreamFlows(ctx, gomock.Any(), gomock.Any()).Return(nil, err)

	otlpLoggerMock := mocks.NewMockOtlpLogger(ctrl)
	otlpLoggerMock.EXPECT().ReceiveFlows(ctx, flowChannelMock).Times(0)

	doneChannel, err := StartLogStreaming(ctx, goldmaneApiMock, otlpLoggerMock, reconnectWaitTime)

	if err == nil {
		t.Errorf("err should not be nil")
	}
	if doneChannel != nil {
		t.Errorf("done channel should be nil since error was encountered")
	}
}
