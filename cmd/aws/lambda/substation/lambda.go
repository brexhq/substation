package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/service"
	"golang.org/x/sync/errgroup"
)

// lambdaAsyncHandler implements ITL that is triggered by an asynchronous invocation
// of the Lambda. Read more about synchronous invocation here:
// https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html.
//
// This implementation of Substation only supports the object data handling pattern
// -- if the payload sent to the Lambda is not JSON, then the invocation will fail.
func lambdaAsyncHandler(ctx context.Context, event map[string]interface{}) error {
	evt, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("lambda async: %v", err)
	}

	sub := cmd.New()

	// retrieve and load configuration
	cfg, err := getConfig(ctx)
	if err != nil {
		return fmt.Errorf("lambda async: %v", err)
	}

	if err := sub.SetConfig(cfg); err != nil {
		return fmt.Errorf("lambda async: %v", err)
	}

	// maintains app state
	group, ctx := errgroup.WithContext(ctx)

	// load
	var sinkWg sync.WaitGroup
	sinkWg.Add(1)
	group.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

	// transform
	var transformWg sync.WaitGroup
	for w := 0; w < sub.Concurrency(); w++ {
		transformWg.Add(1)
		group.Go(func() error {
			return sub.Transform(ctx, &transformWg)
		})
	}

	// ingest
	group.Go(func() error {
		capsule := config.NewCapsule()
		capsule.SetData(evt)

		// do not add metadata -- there is no metadata worth adding from the invocation
		sub.Send(capsule)

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	// block until ITL is complete
	if err := sub.Block(ctx, group); err != nil {
		panic(err)
	}

	return nil
}

// errLambdaSyncMultipleItems is returned when an invocation of the lambdaSync handler
// produces multiple items, which cannot be returned.
const errLambdaSyncMultipleItems = errors.Error("transformed data into multiple items")

// lambdaSyncHandler implements ITL using a request-reply service that is triggered
// by synchronous invocation of the Lambda. Read more about synchronous invocation here:
// https://docs.aws.amazon.com/lambda/latest/dg/invocation-sync.html.
//
// This implementation of Substation has some limitations and requirements:
//
// - Only supports the object data handling pattern -- if the payload sent to the Lambda
// and the result are not JSON, then the invocation will fail
//
// - Only returns a single object -- if many objects may be returned, then they should be
// aggregated into one object using the Aggregate processor
//
// - Must use the gRPC sink configured to send data to localhost:50051 -- data is routed
// from the sink to the handler using the Substation gRPC Sink service
func lambdaSyncHandler(ctx context.Context, event map[string]interface{}) (map[string]interface{}, error) {
	evt, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("lambda sync: %v", err)
	}

	sub := cmd.New()

	// retrieve and load configuration
	cfg, err := getConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("lambda sync: %v", err)
	}

	if err := sub.SetConfig(cfg); err != nil {
		return nil, fmt.Errorf("lambda sync: %v", err)
	}

	// maintains app state
	group, ctx := errgroup.WithContext(ctx)

	// gRPC service, required for catching results from the sink
	server := service.Server{}
	server.Setup()

	// deferring guarantees that the gRPC server will shutdown
	defer server.Stop()

	srv := &service.Sink{}
	server.RegisterSink(srv)

	// gRPC server runs in a goroutine to prevent blocking main
	group.Go(func() error {
		return server.Start("localhost:50051")
	})

	// load
	var sinkWg sync.WaitGroup
	sinkWg.Add(1)
	group.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

	// transform
	var transformWg sync.WaitGroup
	for w := 0; w < sub.Concurrency(); w++ {
		transformWg.Add(1)
		group.Go(func() error {
			return sub.Transform(ctx, &transformWg)
		})
	}

	// ingest
	group.Go(func() error {
		capsule := config.NewCapsule()
		capsule.SetData(evt)

		// do not add metadata -- there is no metadata worth adding from the invocation
		sub.Send(capsule)

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	// block until ITL is complete and the gRPC stream is closed
	if err := sub.Block(ctx, group); err != nil {
		panic(err)
	}
	srv.Block()

	if len(srv.Capsules) > 1 {
		return nil, fmt.Errorf("lambda sync: %v", errLambdaSyncMultipleItems)
	}

	capsule := srv.Capsules[0]
	var output map[string]interface{}
	if err := json.Unmarshal(capsule.Data(), &output); err != nil {
		return nil, fmt.Errorf("lambda sync: %v", err)
	}

	return output, nil
}
