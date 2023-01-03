package sink

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/file"
	pb "github.com/brexhq/substation/proto/v1beta"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// sinkGRPC sinks data to a server that implements the server API for the Sink service.
//
// This sink can be used for inter-process communication (IPC) by using a localhost
// server. By default, the sink creates an insecure connection that is unauthenticated
// and unencrypted.
type sinkGRPC struct {
	// Server is the address and port number for the server that data is sent to.
	Server string `json:"server"`
	// Timeout is the amount of time (in seconds) to wait before cancelling the request.
	//
	// This is optional and defaults to 10 seconds.
	Timeout int `json:"timeout"`
	// Certificate is a file containing a server certificate, which enables SSL/TLS
	// server authentication.
	//
	// This is optional and defaults to unauthenticated and unencrypted connections.
	// The certificate file can be either a path on local disk, an HTTP(S) URL, or
	// an AWS S3 URL.
	Certificate string `json:"certificate"`
}

// Send sinks a channel of encapsulated data with the sink.
func (s *sinkGRPC) Send(ctx context.Context, ch *config.Channel) error {
	// https://grpc.io/docs/guides/auth/#base-case---no-encryption-or-authentication
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())

	// https://grpc.io/docs/guides/auth/#with-server-authentication-ssltls
	if s.Certificate != "" {
		cert, err := file.Get(ctx, s.Certificate)
		if err != nil {
			return fmt.Errorf("sink: grpc: %v", err)
		}
		defer os.Remove(cert)

		c, err := credentials.NewClientTLSFromFile(cert, "")
		if err != nil {
			return fmt.Errorf("sink: grpc: %v", err)
		}

		creds = grpc.WithTransportCredentials(c)
	}

	var opts []grpc.DialOption
	opts = append(opts, creds)

	conn, err := grpc.DialContext(ctx, s.Server, opts...)
	if err != nil {
		return fmt.Errorf("sink: grpc: %v", err)
	}
	defer conn.Close()

	timeout := 10 * time.Second
	if s.Timeout != 0 {
		timeout = time.Duration(s.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client := pb.NewSinkServiceClient(conn)
	stream, err := client.Send(ctx, grpc.WaitForReady(true))
	if err != nil {
		return fmt.Errorf("sink: grpc: %v", err)
	}

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			p := &pb.SendRequest{
				Data: capsule.Data(),
			}

			if err := stream.Send(p); err != nil {
				return fmt.Errorf("sink: grpc: %v", err)
			}
		}
	}

	// server must acknowledge the receipt of all capsules
	// if this doesn't happen, then the app will deadlock
	if _, err := stream.CloseAndRecv(); err != nil {
		return fmt.Errorf("sink: grpc: %v", err)
	}

	return nil
}
