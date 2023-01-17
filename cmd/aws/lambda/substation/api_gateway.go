package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"golang.org/x/sync/errgroup"
)

type gatewayMetadata struct {
	Resource string            `json:"resource"`
	Path     string            `json:"path"`
	Headers  map[string]string `json:"headers"`
}

func gatewayHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sub := cmd.New()

	// retrieve and load configuration
	cfg, err := getConfig(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("gateway: %v", err)
	}

	if err := sub.SetConfig(cfg); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("gateway: %v", err)
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
		if len(request.Body) != 0 {
			capsule := config.NewCapsule()
			capsule.SetData([]byte(request.Body))
			if _, err := capsule.SetMetadata(gatewayMetadata{
				request.Resource,
				request.Path,
				request.Headers,
			}); err != nil {
				return fmt.Errorf("gateway handler: %v", err)
			}

			sub.Send(capsule)
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	// block until ITL is complete
	if err := sub.Block(ctx, group); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("gateway: %v", err)
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}
