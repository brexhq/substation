package service

import (
	"fmt"
	"io"
	"time"

	"github.com/brexhq/substation/config"
	pb "github.com/brexhq/substation/proto"
)

// Sink implements the server API for the Sink service.
type Sink struct {
	pb.UnimplementedSinkServer
	// Capsules can be optionally used to store all capsules sent by the client.
	Capsules []config.Capsule
	// isClosed describes the state of the gRPC stream: false is open and true is closed.
	isClosed bool
}

// Send implements the Send RPC.
func (s *Sink) Send(stream pb.Sink_SendServer) error {
	var count uint32
	capsule := config.NewCapsule()

	// all data is read from the stream before sending an acknowledgement
	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			s.isClosed = true

			return stream.SendAndClose(&pb.Ack{})
		}
		if err != nil {
			return fmt.Errorf("grpc sink recv: %v", err)
		}

		capsule.SetData(recv.Data).SetMetadata(recv.Metadata) //nolint:errcheck // no err check required
		s.Capsules = append(s.Capsules, capsule)

		count++
	}
}

// Block blocks the caller until the gRPC stream is closed. This signals that all data was received by the server.
func (s *Sink) Block() {
	for {
		if !s.isClosed {
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}
}
