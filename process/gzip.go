package process

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
Gzip processes data by compressing or decompressing gzip. The processor supports these patterns:
	data:
		[31 139 8 0 0 0 0 0 0 255 74 203 207 7 4 0 0 255 255 33 101 115 140 3 0 0 0] >>> foo
		foo >>> [31 139 8 0 0 0 0 0 0 255 74 203 207 7 4 0 0 255 255 33 101 115 140 3 0 0 0]

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "gzip",
		"settings": {
			"options": {
				"direction": "from"
			}
		}
	}
*/
type Gzip struct {
	Options   GzipOptions      `json:"options"`
	Condition condition.Config `json:"condition"`
}

/*
GzipOptions contains custom options settings for the Gzip processor:
	Direction:
		the direction of the compression
		must be one of:
			to: compress data to gzip
			from: decompress data from gzip
*/
type GzipOptions struct {
	Direction string `json:"direction"`
}

func (p Gzip) from(data []byte) ([]byte, error) {
	r := bytes.NewReader(data)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip: %v", err)
	}

	output, err := ioutil.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("gzip: %v", err)
	}

	return output, nil
}

func (p Gzip) to(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("gzip: %v", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("gzip: %v", err)
	}

	return buf.Bytes(), nil
}

// ApplyBatch processes a slice of encapsulated data with the Gzip processor. Conditions are optionally applied to the data to enable processing.
func (p Gzip) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process gzip applybatch: %v", err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("process gzip applybatch: %v", err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Gzip processor.
func (p Gzip) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return cap, fmt.Errorf("process gzip apply: options %+v: %v", p.Options, processorMissingRequiredOptions)
	}

	var value []byte
	switch p.Options.Direction {
	case "from":
		from, err := p.from(cap.Data())
		if err != nil {
			return cap, fmt.Errorf("process gzip apply: %v", err)
		}

		value = from
	case "to":
		to, err := p.to(cap.Data())
		if err != nil {
			return cap, fmt.Errorf("process gzip apply: %v", err)
		}

		value = to
	default:
		return cap, fmt.Errorf("process gzip apply: direction %s: %v", p.Options.Direction, processorInvalidDirection)
	}

	cap.SetData(value)
	return cap, nil
}
