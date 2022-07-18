package process

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"

	"github.com/brexhq/substation/condition"
)

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

/*
Gzip processes data by compressing or decompressing gzip. The processor supports these patterns:
	data:
		[31 139 8 0 0 0 0 0 0 255 74 203 207 7 4 0 0 255 255 33 101 115 140 3 0 0 0] >>> foo
		foo >>> [31 139 8 0 0 0 0 0 0 255 74 203 207 7 4 0 0 255 255 33 101 115 140 3 0 0 0]

The processor uses this Jsonnet configuration:
	{
		type: 'gzip',
		settings: {
			options: {
				direction: 'from',
			},
		},
	}
*/
type Gzip struct {
	Options   GzipOptions              `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
}

// Slice processes a slice of bytes with the Gzip processor. Conditions are optionally applied on the bytes to enable processing.
func (p Gzip) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("slicer: %v", err)
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

// Byte processes bytes with the Gzip processor.
func (p Gzip) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	switch s := p.Options.Direction; s {
	case "from":
		tmp, err := p.from(data)
		if err != nil {
			return nil, fmt.Errorf("byter settings %+v: %w", p, err)
		}

		return tmp, nil
	case "to":
		tmp, err := p.to(data)
		if err != nil {
			return nil, fmt.Errorf("byter settings %+v: %w", p, err)
		}

		return tmp, nil
	}

	return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
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
