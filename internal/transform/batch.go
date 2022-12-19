package transform

import (
	"context"
	"time"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/metrics"
	"github.com/brexhq/substation/process"
)

/*
Batch transforms data by applying a series of processors to a slice of encapsulated data. Data processing is iterative and each processor is enabled through conditions.

Below is an example that shows how a single JSON object is iteratively modified through this transform:

	{"hello":"world"} // input event
	{"hello":"world","foo":"bar"} // insert value "bar" into key "foo"
	{"hello":"world","foo":"bar","baz":"qux"} // insert value "qux" into key "bar"
	{"hello":"world","foo":"bar.qux"} // concat values from "foo" and "baz" into key "foo" with separator "."

When loaded with a factory, the transform uses this JSON configuration:

	{
		"type": "batch",
		"processors": [
			{
				"type": "hash",
				"settings": {
					"condition": {
						"inspectors": [ ],
						"operator": ""
					},
					"input_key": "@this",
					"output_key": "event.hash"
					"options": {
						"algorithm": "sha256"
					}
				}
			}
		]
	}
*/
type Batch struct {
	Processors []config.Config `json:"processors"`
}

// Transform processes a channel of encapsulated data with the Batch transform.
func (transform *Batch) Transform(ctx context.Context, in, out *config.Channel) error {
	batchers, err := process.MakeBatchers(transform.Processors...)
	if err != nil {
		return err
	}

	//nolint: errcheck // errors are ignored in case processing fails in a single applicator
	defer process.CloseBatchers(ctx, batchers...)

	var received int
	// read encapsulated data from the input channel into a batch
	batch := make([]config.Capsule, 0, 10)
	for capsule := range in.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			batch = append(batch, capsule)
			received++

			// sleep load balances data across transform goroutines, otherwise a single goroutine may receive more capsules than others
			// this creates 1 second of delay for every 100,000 capsules put into the input channel
			time.Sleep(time.Duration(10) * time.Microsecond)
		}
	}

	// iteratively process the batch of encapsulated data
	batch, err = process.Batch(ctx, batch, batchers...)
	if err != nil {
		return err
	}

	var sent int
	// write the processed, encapsulated data to the output channel
	for _, capsule := range batch {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			out.Send(capsule)
			sent++
		}
	}

	_ = metrics.Generate(ctx, metrics.Data{
		Name:  "CapsulesReceived",
		Value: received,
	})

	_ = metrics.Generate(ctx, metrics.Data{
		Name:  "CapsulesSent",
		Value: sent,
	})

	return nil
}
