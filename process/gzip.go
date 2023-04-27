package process

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// gzip processes data by compressing or decompressing gzip.
//
// This processor supports the data handling pattern.
type procGzip struct {
	process
	Options procGzipOptions `json:"options"`
}

type procGzipOptions struct {
	// Direction determines whether data is compressed or decompressed.
	//
	// Must be one of:
	//	- to: compress to gzip
	// 	- from: decompress from gzip
	Direction string `json:"direction"`
}

// Create a new gzip processor.
func newProcGzip(ctx context.Context, cfg config.Config) (p procGzip, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procGzip{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procGzip{}, err
	}

	// fail for invalid option.direction
	if !slices.Contains(
		[]string{"to", "from"},
		p.Options.Direction) {
		return procGzip{}, fmt.Errorf("process: gzip: direction %q: %v", p.Options.Direction, errors.ErrInvalidOption)
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procGzip) String() string {
	return toString(p)
}

func (p procGzip) from(data []byte) ([]byte, error) {
	r := bytes.NewReader(data)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("process: gzip: %v", err)
	}

	output, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("process: gzip: %v", err)
	}

	return output, nil
}

func (p procGzip) to(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("process: gzip: %v", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("process: gzip: %v", err)
	}

	return buf.Bytes(), nil
}

// Closes resources opened by the processor.
func (p procGzip) Close(context.Context) error {
	return nil
}

// Stream processes a pipeline of capsules with the processor.
func (p procGzip) Stream(ctx context.Context, in, out *config.Channel) error {
	return streamApply(ctx, in, out, p)
}

// Batch processes one or more capsules with the processor.
func (p procGzip) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p)
}

// Apply processes a capsule with the processor.
func (p procGzip) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	if ok, err := p.operator.Operate(ctx, capsule); err != nil {
		return capsule, fmt.Errorf("process: gzip: %v", err)
	} else if !ok {
		return capsule, nil
	}

	var value []byte
	switch p.Options.Direction {
	case "from":
		from, err := p.from(capsule.Data())
		if err != nil {
			return capsule, fmt.Errorf("process: gzip: %v", err)
		}

		value = from
	case "to":
		to, err := p.to(capsule.Data())
		if err != nil {
			return capsule, fmt.Errorf("process: gzip: %v", err)
		}

		value = to
	default:
		return capsule, fmt.Errorf("process: gzip: direction %s: %v", p.Options.Direction, errInvalidDirection)
	}

	capsule.SetData(value)
	return capsule, nil
}
