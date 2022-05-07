package transform

import (
	"context"

	"github.com/brexhq/substation/process"
)

/*
Process transforms data by applying processors. Each processor is enabled through conditions. This transform uses process Slicers to iteratively modify slices of bytes.

Below is an example that shows how a single JSON object is iteratively modified through this transform:
	{"hello":"world"} // input event
	{"hello":"world","foo":"bar"} // insert value "bar" into key "foo"
	{"hello":"world","foo":"bar","baz":"qux"} // insert value "qux" into key "bar"
	{"hello":"world","foo":"bar.qux"} // concat vaues from "foo" and "baz" into key "foo" with separator "."

The transform uses this Jsonnet configuration:
	{
		type: 'process',
		processors: [
			{
				"settings": {
					"condition": {
						"inspectors": [ ],
						"operator": ""
					},
					"input": {
						"key": "@this"
					},
					"options": {
						"algorithm": "sha256"
					},
					"output": {
						"key": "event.hash"
					}
				},
				"type": "hash"
			},
		]
	}
*/
type Process struct {
	Processors []process.Config `mapstructure:"processors"`
}

// Transform processes a channel of bytes with the Process transform.
func (transform *Process) Transform(ctx context.Context, in <-chan []byte, out chan<- []byte, kill chan struct{}) error {
	slicers, err := process.MakeAllSlicers(transform.Processors)
	if err != nil {
		return err
	}

	slice := make([][]byte, 0, 100)
	for data := range in {
		select {
		case <-kill:
			return nil
		default:
			slice = append(slice, data)
		}
	}

	slice, err = process.Slice(ctx, slicers, slice)
	if err != nil {
		return err
	}

	for _, data := range slice {
		select {
		case <-kill:
			return nil
		default:
			out <- data
		}
	}

	return nil
}
