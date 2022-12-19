package process

import (
	"bytes"
	gogzip "compress/gzip"
	"context"
	"fmt"
	"io"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

type gzip struct {
	Options   gzipOptions      `json:"options"`
	Condition condition.Config `json:"condition"`
}

type gzipOptions struct {
	Direction string `json:"direction"`
}

func (p gzip) from(data []byte) ([]byte, error) {
	r := bytes.NewReader(data)
	gz, err := gogzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("process gzip: %v", err)
	}

	output, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("process gzip: %v", err)
	}

	return output, nil
}

func (p gzip) to(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gogzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("process gzip: %v", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("process gzip: %v", err)
	}

	return buf.Bytes(), nil
}

// Close closes resources opened by the gzip processor.
func (p gzip) Close(context.Context) error {
	return nil
}

func (p gzip) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process gzip: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the gzip processor.
func (p gzip) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
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
