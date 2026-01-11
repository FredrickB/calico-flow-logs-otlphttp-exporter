package otlp

import (
	"context"
	"errors"
	"sync"
	"testing"

	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	flow = pb.Flow{
		SourceLabels: []string{"test=true"},
	}
	processedFlows = []string{}
)

type FakeProcessor struct {
	ProcessedFlows []string
}

func (processor *FakeProcessor) Log(_ context.Context, message string) {
	processor.ProcessedFlows = append(processor.ProcessedFlows, message)
}

func (processor *FakeProcessor) Close(_ context.Context) error {
	return nil
}

func TestReceiveFlows(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flowCh := make(chan *pb.Flow)

	processorFake := FakeProcessor{ProcessedFlows: processedFlows}

	flowJson, marshalErr := protojson.Marshal(&flow)
	if marshalErr != nil {
		t.Errorf("should not fail to marshal flow")
	}

	logger := NewLogger(&processorFake)

	go func() {
		logger.ReceiveFlows(ctx, flowCh)
	}()

	flowCh <- &flow

	actualFlow := processorFake.ProcessedFlows[0]

	if actualFlow != string(flowJson) {
		t.Errorf("expected: %s, actual: %s", string(flowJson), actualFlow)
	}
}

func TestReceiveFlowsWithContextCancelDoesNotInvokeCallToProcessor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	processorFake := FakeProcessor{ProcessedFlows: processedFlows}

	logger := NewLogger(&processorFake)
	flowCh := make(chan *pb.Flow)

	var wg sync.WaitGroup
	wg.Add(1)

	var err error

	go func() {
		err = logger.ReceiveFlows(ctx, flowCh)
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
	processorFake := FakeProcessor{ProcessedFlows: processedFlows}

	logger := NewLogger(&processorFake)
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
