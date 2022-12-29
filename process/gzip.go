package process

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"

	"github.com/brexhq/substation/config"
)

// gzip processes data by compressing or decompressing gzip.
//
// This processor supports the data handling pattern.
type _gzip struct {
	process
	Options _gzipOptions `json:"options"`
}

type _gzipOptions struct {
	// Direction determines whether data is compressed or decompressed.
	//
	// Must be one of:
	//	- to: compress to gzip
	// 	- from: decompress from gzip
	Direction string `json:"direction"`
}

// String returns the processor settings as an object.
func (p _gzip) String() string {
	return toString(p)
}

func (p _gzip) from(data []byte) ([]byte, error) {
	r := bytes.NewReader(data)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("process _gzip: %v", err)
	}

	output, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("process _gzip: %v", err)
	}

	return output, nil
}

func (p _gzip) to(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("process _gzip: %v", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("process _gzip: %v", err)
	}

	return buf.Bytes(), nil
}

// Close closes resources opened by the processor.
func (p _gzip) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _gzip) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
func (p _gzip) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return capsule, fmt.Errorf("process _gzip: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	var value []byte
	switch p.Options.Direction {
	case "from":
		from, err := p.from(capsule.Data())
		if err != nil {
			return capsule, fmt.Errorf("process _gzip: %v", err)
		}

		value = from
	case "to":
		to, err := p.to(capsule.Data())
		if err != nil {
			return capsule, fmt.Errorf("process _gzip: %v", err)
		}

		value = to
	default:
		return capsule, fmt.Errorf("process _gzip: direction %s: %v", p.Options.Direction, errInvalidDirection)
	}

	capsule.SetData(value)
	return capsule, nil
}
