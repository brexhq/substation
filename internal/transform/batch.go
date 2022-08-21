package transform

import (
	"context"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/process"
)

/*
Batch transforms data by applying a series of processors to a slice of encapsulated data. Data processing is iterative and each processor is enabled through conditions.

Below is an example that shows how a single JSON object is iteratively modified through this transform:
	{"hello":"world"} // input event
	{"hello":"world","foo":"bar"} // insert value "bar" into key "foo"
	{"hello":"world","foo":"bar","baz":"qux"} // insert value "qux" into key "bar"
	{"hello":"world","foo":"bar.qux"} // concat vaues from "foo" and "baz" into key "foo" with separator "."

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
func (transform *Batch) Transform(ctx context.Context, in <-chan config.Capsule, out chan<- config.Capsule, kill chan struct{}) error {
	applicators, err := process.MakeBatchApplicators(transform.Processors)
	if err != nil {
		return err
	}

	// read encapsulated data from the input channel into a batch
	batch := make([]config.Capsule, 0, 10)
	for cap := range in {
		select {
		case <-kill:
			return nil
		default:
			batch = append(batch, cap)
		}
	}

	// iteratively process the batch of encapsulated data
	batch, err = process.ApplyBatch(ctx, batch, applicators...)
	if err != nil {
		return err
	}

	// write the processed, encapsulated data to the output channel
	// if a signal is received on the kill channel, then this is interrupted
	for _, cap := range batch {
		select {
		case <-kill:
			return nil
		default:
			out <- cap
		}
	}

	return nil
}