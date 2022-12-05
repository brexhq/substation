package process

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"

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
		direction of the compression
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
		return nil, fmt.Errorf("process gzip: %v", err)
	}

	output, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("process gzip: %v", err)
	}

	return output, nil
}

func (p Gzip) to(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("process gzip: %v", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("process gzip: %v", err)
	}

	return buf.Bytes(), nil
}

// Close closes resources opened by the Gzip processor.
func (p Gzip) Close(context.Context) error {
	return nil
}

// ApplyBatch processes a slice of encapsulated data with the Gzip processor. Conditions are optionally applied to the data to enable processing.
func (p Gzip) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process gzip: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process gzip: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Gzip processor.
func (p Gzip) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return capsule, fmt.Errorf("process gzip: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	var value []byte
	switch p.Options.Direction {
	case "from":
		from, err := p.from(capsule.Data())
		if err != nil {
			return capsule, fmt.Errorf("process gzip: %v", err)
		}

		value = from
	case "to":
		to, err := p.to(capsule.Data())
		if err != nil {
			return capsule, fmt.Errorf("process gzip: %v", err)
		}

		value = to
	default:
		return capsule, fmt.Errorf("process gzip: direction %s: %v", p.Options.Direction, errInvalidDirection)
	}

	capsule.SetData(value)
	return capsule, nil
}
