package goldmane

import (
	"context"
	"errors"
	"io"
	"log"
	"time"

	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type GoldmaneApi interface {
	StreamFlows(context context.Context, done chan<- bool, reconnectWaitTime time.Duration) (<-chan *pb.Flow, error)
}

type GoldmaneClient struct {
	client pb.FlowsClient
}

func NewClient(host string, client pb.FlowsClient) *GoldmaneClient {
	return &GoldmaneClient{client: client}
}

func (c *GoldmaneClient) Close() error {
	return nil
}

func (c *GoldmaneClient) StreamFlows(context context.Context, done chan<- bool, reconnectWaitTime time.Duration) (<-chan *pb.Flow, error) {
	data := make(chan *pb.Flow)

	go func() {
		for {
			select {
			case <-context.Done():
				log.Println("Context done, terminating")
				return
			default:
				stream, err := c.client.Stream(context, &pb.FlowStreamRequest{})

				if err != nil {
					log.Printf("Failed to create stream: %s. Sleeping for %s", err, reconnectWaitTime)
					time.Sleep(reconnectWaitTime)
					continue
				}

				log.Printf("Stream created")

				// block until streaming fails, then check return
				// value to see if reconnect should be done or if
				// the execution should be terminated
				reconnect, _ := streamFlowsUntilError(context, stream, data)
				if !reconnect {
					close(data)
					done <- true
					return
				}
			}
		}
	}()

	return data, nil
}

// start streaming flow logs to channel `data` until error is
// received from grpc. If there is a possibillity for recovery
// by reconnecting, flag it to the caller using `reconnect`
func streamFlowsUntilError(context context.Context, stream grpc.ServerStreamingClient[pb.FlowResult], data chan<- *pb.Flow) (reconnect bool, err error) {
	// start out with the assumption that
	// reconnecting won't be done for the
	// majority of errors
	reconnect = false

	log.Printf("Streaming logs...")

	for {
		select {
		case <-context.Done():
			log.Println("Context done, terminating")
			return
		default:
			var flowResult pb.FlowResult
			err = stream.RecvMsg(&flowResult)

			if errors.Is(err, io.EOF) {
				// the sender has nothing more to send, terminate
				log.Printf("Received EOF, terminating")
				return
			}

			if err != nil {
				if status, ok := status.FromError(err); ok {
					// known error, reconnect
					log.Printf("Status received: %s, trigger reconnect", status.Code())
					reconnect = true
				} else {
					// unrecognizeable error, terminate
					log.Printf("Unknown error: %s, terminating", err)
				}
				return
			}

			data <- flowResult.Flow
		}
	}
}
