package util

import (
	"context"
	"errors"
	"testing"
	"time"

	mocks "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/mocks"
	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/protos"

	"go.uber.org/mock/gomock"
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

	flowChannelMock := make(<-chan *pb.Flow)
	reconnectWaitTime := 2 * time.Second

	goldmaneApiMock := mocks.NewMockGoldmaneApi(ctrl)
	goldmaneApiMock.EXPECT().StreamFlows(context.TODO(), gomock.Any(), gomock.Any()).Return(flowChannelMock, nil)

	otlpLoggerMock := mocks.NewMockOtlpLogger(ctrl)
	otlpLoggerMock.EXPECT().ReceiveFlows(context.TODO(), flowChannelMock).Times(1)

	doneChannel, err := StartLogStreaming(context.TODO(), goldmaneApiMock, otlpLoggerMock, reconnectWaitTime)

	if err != nil {
		t.Errorf("err should be nil")
	}
	if doneChannel == nil {
		t.Errorf("done channel should not be nil")
	}
}

func TestStartLogStreamingFailsIfStreamFlowsReturnsErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowChannelMock := make(<-chan *pb.Flow)
	err := errors.New("mocked error")
	reconnectWaitTime := 2 * time.Second

	goldmaneApiMock := mocks.NewMockGoldmaneApi(ctrl)
	goldmaneApiMock.EXPECT().StreamFlows(context.TODO(), gomock.Any(), gomock.Any()).Return(nil, err)

	otlpLoggerMock := mocks.NewMockOtlpLogger(ctrl)
	otlpLoggerMock.EXPECT().ReceiveFlows(context.TODO(), flowChannelMock).Times(0)

	doneChannel, err := StartLogStreaming(context.TODO(), goldmaneApiMock, otlpLoggerMock, reconnectWaitTime)

	if err == nil {
		t.Errorf("err should not be nil")
	}
	if doneChannel != nil {
		t.Errorf("done channel should be nil since error was encountered")
	}
}
