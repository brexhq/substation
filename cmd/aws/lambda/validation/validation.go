package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/brexhq/substation/cmd"
	"golang.org/x/sync/errgroup"
)

func main() {
	lambda.Start(handler)
}

type validationEvent struct {
	Content string `json:"content"`
	URI     string `json:"uri"`
}

func handler(ctx context.Context, event json.RawMessage) error {
	var e validationEvent
	err := json.Unmarshal(event, &e)
	if err != nil {
		return err
	}

	cfg, err := base64.RawStdEncoding.DecodeString(e.Content)
	if err != nil {
		return err
	}

	sub := cmd.New()
	if err := sub.SetConfig(bytes.NewReader(cfg)); err != nil {
		return fmt.Errorf("validation: %v", err)
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
	transformWg.Add(1)
	group.Go(func() error {
		return sub.Transform(ctx, &transformWg)
	})

	// ingest nothing
	group.Go(func() error {
		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	// block until ITL is complete
	if err := sub.Block(ctx, group); err != nil {
		return fmt.Errorf("validation: %v", err)
	}

	return nil
}
