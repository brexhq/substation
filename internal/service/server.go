package service

import (
	"fmt"
	"net"

	pb "github.com/brexhq/substation/proto"
	"google.golang.org/grpc"
)

// New returns a configured gRPC server.
func New(opt ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(opt...)
}

// Server wraps a gRPC server and provides methods for managing server state.
type Server struct {
	server *grpc.Server
}

// Setup creates a new gRPC server.
func (s *Server) New(opt ...grpc.ServerOption) {
	s.server = New(opt...)
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

// RegisterSink registers the server API for the Sink service with the gRPC server.
func (s *Server) RegisterSink(srv *Sink) {
	pb.RegisterSinkServer(s.server, srv)
}
