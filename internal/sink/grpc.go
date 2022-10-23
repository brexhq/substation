package sink

import (
	"context"
	"fmt"
	"os"
	"time"

	pb "github.com/brexhq/substation"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

/*
gRPC sinks data to a server that implements the Substation Sink rpc. This sink can also be used for inter-process communication (IPC) by using a localhost server. By default, the sink creates an insecure connection that is unauthenticated and unencrypted.

The sink has these settings:

	Server:
		Address and port number for the server that data is sent to
	Timeout (optional):
		Amount of time (in seconds) to wait before cancelling the request
		defaults to 10 seconds
	Certificate (optional):
		File containing the server certificate, enables SSL/TLS server authentication
		The certificate file can be stored locally or remotely

When loaded with a factory, the sink uses this JSON configuration:

	{
		"type": "grpc",
		"settings": {
			"server": "localhost:50051"
		}
	}
*/
type Grpc struct {
	Server      string `json:"server"`
	Timeout     int    `json:"timeout"`
	Certificate string `json:"certificate"`
}

func (sink *Grpc) Send(ctx context.Context, ch *config.Channel) error {
	// https://grpc.io/docs/guides/auth/#base-case---no-encryption-or-authentication
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())

	// https://grpc.io/docs/guides/auth/#with-server-authentication-ssltls
	if sink.Certificate != "" {
		cert, err := file.Get(ctx, sink.Certificate)
		if err != nil {
			return fmt.Errorf("sink grpc: %v", err)
		}
		defer os.Remove(cert)

		c, err := credentials.NewClientTLSFromFile(cert, "")
		if err != nil {
			return fmt.Errorf("sink grpc: %v", err)
		}

		creds = grpc.WithTransportCredentials(c)
	}

	var opts []grpc.DialOption
	opts = append(opts, creds)

	conn, err := grpc.DialContext(ctx, sink.Server, opts...)
	if err != nil {
		return fmt.Errorf("sink grpc: %v", err)
	}
	defer conn.Close()

	timeout := 10 * time.Second
	if sink.Timeout != 0 {
		timeout = time.Duration(sink.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client := pb.NewSubstationClient(conn)
	stream, err := client.Sink(ctx, grpc.WaitForReady(true))
	if err != nil {
		return fmt.Errorf("sink grpc: %v", err)
	}

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			p := &pb.Capsule{
				Data: capsule.Data(),
			}

			if err := stream.Send(p); err != nil {
				return fmt.Errorf("sink grpc: %v", err)
			}
		}
	}

	// server must acknowledge the receipt of all capsules
	// if this doesn't happen, then the app will deadlock
	if _, err := stream.CloseAndRecv(); err != nil {
		return fmt.Errorf("sink grpc: %v", err)
	}

	return nil
}
