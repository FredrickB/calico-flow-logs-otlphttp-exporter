package otlp

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/mocks"
	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/protos"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	flow = pb.Flow{
		SourceLabels: []string{"test=true"},
	}
)

func TestReceiveFlows(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	flowChannel := make(chan *pb.Flow, 2)
	flowChannel <- &flow

	processorMock := mocks.NewMockProcessor(ctrl)
	flowJson, marshalErr := protojson.Marshal(&flow)
	if marshalErr != nil {
		t.Errorf("should not fail to marshal flow protobuf")
	}
	processorMock.EXPECT().Log(ctx, string(flowJson)).Times(1)

	logger := NewLogger(processorMock)

	var err error

	go func() {
		wg.Done()
		err = logger.ReceiveFlows(ctx, flowChannel)
	}()

	wg.Wait()

	if err != nil {
		t.Errorf("err should be nil")
	}
}

func TestReceiveFlowsWithContextCancelDoesNotInvokeCallToProcessor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())

	processorMock := mocks.NewMockProcessor(ctrl)
	processorMock.EXPECT().Log(gomock.Any(), gomock.Any()).Times(0)

	logger := NewLogger(processorMock)
	flowChannel := make(chan *pb.Flow)

	var wg sync.WaitGroup
	wg.Add(1)

	var err error

	go func() {
		err = logger.ReceiveFlows(ctx, flowChannel)
		wg.Done()
	}()

	cancel()

	// below this pont, err should be set
	wg.Wait()

	if !errors.Is(err, ErrContextCancelled) {
		t.Errorf("should have returned expected error")
	}
}

func TestReceiveFlowsWithClosedDataChannelDoesNotInvokeCallToProcessor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	processorMock := mocks.NewMockProcessor(ctrl)
	processorMock.EXPECT().Log(gomock.Any(), gomock.Any()).Times(0)

	logger := NewLogger(processorMock)
	flowChannel := make(chan *pb.Flow)

	var wg sync.WaitGroup
	wg.Add(1)

	var err error

	go func() {
		err = logger.ReceiveFlows(context.Background(), flowChannel)
		wg.Done()
	}()

	close(flowChannel)

	// below this pont, err should be set
	wg.Wait()

	if !errors.Is(err, ErrDataChannelClosed) {
		t.Errorf("should have returned expected error")
	}
}
