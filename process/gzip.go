package process

import (
	"bytes"
	"compress/gzip"
	"context"
	"io/ioutil"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
)

// GzipInvalidSettings is returned when the Gzip processor is configured with invalid Input and Output settings.
const GzipInvalidSettings = errors.Error("GzipInvalidSettings")

// GzipInvalidDirection is used when an invalid direction is given to the processor
const GzipInvalidDirection = errors.Error("GzipInvalidDirection")

/*
GzipOptions contains custom options settings for the Gzip processor:
	Direction:
		the direction of the compression
		must be one of:
			to: compress data to gzip
			from: decompress data from gzip
*/
type GzipOptions struct {
	Direction string `mapstructure:"direction"`
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
			direction: 'from',
		},
	}
*/
type Gzip struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Options   GzipOptions              `mapstructure:"options"`
}

// Channel processes a data channel of byte slices with the Gzip processor. Conditions are optionally applied on the channel data to enable processing.
func (p Gzip) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	var array [][]byte
	for data := range ch {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			array = append(array, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
		array = append(array, processed)
	}

	output := make(chan []byte, len(array))
	for _, x := range array {
		output <- x
	}
	close(output)
	return output, nil
}

// Byte processes a byte slice with the Gzip processor.
func (p Gzip) Byte(ctx context.Context, data []byte) ([]byte, error) {
	switch s := p.Options.Direction; s {
	case "from":
		tmp, err := p.from(data)
		if err != nil {
			return nil, err
		}

		return tmp, nil
	case "to":
		tmp, err := p.to(data)
		if err != nil {
			return nil, err
		}

		return tmp, nil
	default:
		return nil, GzipInvalidDirection
	}
}

func (p Gzip) from(data []byte) ([]byte, error) {
	r := bytes.NewReader(data)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	output, err := ioutil.ReadAll(gz)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (p Gzip) to(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
