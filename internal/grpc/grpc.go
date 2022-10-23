package grpc

import (
	"fmt"
	"io"
	"net"
	"time"

	pb "github.com/brexhq/substation"
	"github.com/brexhq/substation/config"
	"google.golang.org/grpc"
)

// Server wraps a gRPC server.
type Server struct {
	server *grpc.Server
}

// New creates a new gPRC server.
func (s *Server) New(opt ...grpc.ServerOption) {
	s.server = grpc.NewServer(opt...)
}

// Start starts the gRPC server. This method blocks the caller until the server is stopped.
func (s *Server) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("grpc start: %v", err)
	}

	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("grpc serve: %v", err)
	}

	return nil
}

// Stop stops the gRPC server.
func (s *Server) Stop() {
	s.server.Stop()
}

// Register registers a Substation gRPC service with the server.
func (s *Server) Register(srv *Service) {
	pb.RegisterSubstationServer(s.server, srv)
}

// Service implements the Substation gRPC service.
type Service struct {
	pb.UnimplementedSubstationServer
	// Capsules can be optionally used to store all capsules sent by the client in client-side and bidirectional streaming RPCs.
	Capsules []config.Capsule
	// EOF describes the state of a gRPC stream: false is open and true is closed.
	EOF bool
}

// Sink implements the Substation gRPC Sink rpc.
func (s *Service) Sink(stream pb.Substation_SinkServer) error {
	var count uint32
	capsule := config.NewCapsule()

	// all data is read from the stream before sending an acknowledgement
	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			s.EOF = true

			return stream.SendAndClose(&pb.Ack{
				Count: count,
			})
		}
		if err != nil {
			return fmt.Errorf("grpc sink recv: %v", err)
		}

		capsule.SetData(recv.Data).SetMetadata(recv.Metadata) //nolint:errcheck // no err check required
		s.Capsules = append(s.Capsules, capsule)

		count++
	}
}

// Block blocks the caller until a gRPC stream is closed. This can be used by client-side and bidirectional streaming RPCs to signal that all data was received.
func (s *Service) Block() {
	for {
		if !s.EOF {
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}
}
