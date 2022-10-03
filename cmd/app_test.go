package cmd

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/brexhq/substation/config"
	"go.uber.org/goleak"
	"golang.org/x/sync/errgroup"
)

var appLeaksTest = []struct {
	name   string
	config []byte
}{
	{
		"invalid sink",
		[]byte(`
		{
			"sink": {
			   "type": "fooer"
			},
			"transform": {
			   "type": "transfer"
			}
		 }
		 `),
	},
	{
		"invalid transform",
		[]byte(`
		{
			"sink": {
			   "type": "stdout"
			},
			"transform": {
			   "type": "fooer"
			}
		 }
		 `),
	},
	{
		"invalid processor",
		[]byte(`
		{
			"sink": {
				"type": "stdout"
			},
			"transform": {
				"settings": {
				   "processors": [
					  {
						 "type": "fooer"
					  }
				   ]
				},
				"type": "batch"
			 }
		 }
		 `),
	},
	{
		"invalid processor settings",
		[]byte(`
		{
			"sink": {
				"type": "stdout"
			},
			"transform": {
				"settings": {
				   "processors": [
					  {
						 "settings": {},
						 "type": "copy"
					  }
				   ]
				},
				"type": "batch"
			 }
		 }
		 `),
	},
	{
		"valid config",
		[]byte(`
		{
			"sink": {
				"type": "stdout"
			},
			"transform": {
				"settings": {
				   "processors": [
					{
						"settings": {
						   "input_key": "foo",
						   "output_key": "baz"
						},
						"type": "copy"
					 }		 
				   ]
				},
				"type": "batch"
			 }
		 }
		 `),
	},
}

// TestAppLeaks contains a fully functional application and tests multiple configurations for goroutine leaks.
func TestAppLeaks(t *testing.T) {
	defer goleak.VerifyNone(t)

	for _, test := range appLeaksTest {
		sub := New()
		json.Unmarshal(test.config, &sub.Config)

		group, ctx := errgroup.WithContext(context.TODO())

		var sinkWg sync.WaitGroup
		sinkWg.Add(1)
		group.Go(func() error {
			return sub.Sink(ctx, &sinkWg)
		})

		var transformWg sync.WaitGroup
		for w := 0; w < sub.Concurrency(); w++ {
			transformWg.Add(1)
			group.Go(func() error {
				return sub.Transform(ctx, &transformWg)
			})
		}

		// ingest
		group.Go(func() error {
			cap := config.NewCapsule()
			cap.SetData([]byte(`{"foo":"bar"}`))

			for w := 0; w < 10; w++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					sub.Send(cap)
				}
			}

			sub.WaitTransform(&transformWg)
			sub.WaitSink(&sinkWg)

			return nil
		})

		// block without checking for errors
		// this test only checks for leaks
		sub.Block(ctx, group)
	}
}
