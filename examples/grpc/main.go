package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/grpc"
	"golang.org/x/sync/errgroup"
)

func main() {
	sub := cmd.New()

	bytes, err := os.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bytes, &sub.Config); err != nil {
		panic(err)
	}

	// maintains app state
	group, ctx := errgroup.WithContext(context.TODO())

	// create the gRPC server, the gRPC service, and register the service with the server
	server := grpc.Server{}
	server.New()
	// deferring guarantees that the gRPC server will shutdown
	defer server.Stop()

	srv := &grpc.Service{}
	server.Register(srv)

	// gRPC server runs in a goroutine to prevent blocking main
	group.Go(func() error {
		return server.Start("localhost:50051")
	})

	// sink goroutine
	var sinkWg sync.WaitGroup
	sinkWg.Add(1)
	group.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

	// transform goroutine
	var transformWg sync.WaitGroup
	transformWg.Add(1)
	group.Go(func() error {
		return sub.Transform(ctx, &transformWg)
	})

	// ingest goroutine
	group.Go(func() error {
		data := [][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
			[]byte(`{"qux":"corge"}`),
		}

		cap := config.NewCapsule()

		fmt.Println("sending capsules into Substation ...")
		for _, d := range data {
			fmt.Println(string(d))
			cap.SetData(d)
			sub.Send(cap)
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	// block until all Substation processing is complete
	if err := sub.Block(ctx, group); err != nil {
		panic(err)
	}

	// block until the gRPC server has received all capsules and the stream is closed
	srv.Block()

	fmt.Println("returning capsules stored in gRPC service ...")
	for _, cap := range srv.Capsules {
		fmt.Println(string(cap.Data()))
	}
}
