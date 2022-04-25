package process

import (
	"bytes"
	"compress/gzip"
	"context"
	"io/ioutil"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
)

// GzipInvalidDirection is used when an invalid direction is given to the processor
const GzipInvalidDirection = errors.Error("GzipInvalidDirection")

/*
GzipOptions contain custom options settings for this processor.

Direction: the direction of the compression, either to (compress) or from (decompress) Gzip.
*/
type GzipOptions struct {
	Direction string `mapstructure:"direction"`
}

// Gzip implements the Byter and Channeler interfaces and converts bytes to and from Gzip. More information is available in the README.
type Gzip struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Options   GzipOptions              `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Gzip) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	var array [][]byte

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

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

// Byte processes a byte slice with this processor.
func (p Gzip) Byte(ctx context.Context, data []byte) ([]byte, error) {
	if p.Options.Direction == "from" {
		tmp, err := fromGzip(data)
		if err != nil {
			return nil, err
		}

		return tmp, nil
	} else if p.Options.Direction == "to" {
		tmp, err := toGzip(data)
		if err != nil {
			return nil, err
		}

		return tmp, nil
	} else {
		return nil, GzipInvalidDirection
	}
}

func fromGzip(data []byte) ([]byte, error) {
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

func toGzip(data []byte) ([]byte, error) {
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
