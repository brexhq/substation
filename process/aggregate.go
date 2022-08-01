package process

import (
	"bytes"
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	"github.com/jshlbrd/go-aggregate"
)

// AggregateBufferSizeLimit is returned when the aggregate's buffer size limit is reached. If this error occurs, then increase the size of the buffer or use the Drop processor to remove data that exceeds the buffer limit.
const AggregateBufferSizeLimit = errors.Error("AggregateBufferSizeLimit")

/*
AggregateOptions contains custom options settings for the Aggregate processor:
	AggregateKey (optional):
		the JSON key-value that is used to organize aggregated data
		defaults to empty string, only applies to JSON
	Separator (optional):
		the string that separates aggregated data
		defaults to empty string, only applies to data
	MaxCount (optional):
		the maximum number of items stored in a buffer when aggregating data
		defaults to 1000
	MaxSize (optional):
		the maximum size, in bytes, of items stored in a buffer when aggregating data
		defaults to 10000 (10KB)
*/
type AggregateOptions struct {
	AggregateKey string `json:"aggregate_key"`
	Separator    string `json:"separator"`
	MaxCount     int    `json:"max_count"`
	MaxSize      int    `json:"max_size"`
}

/*
Aggregate processes data by buffering and aggregating it
into a single item.

Data is processed by aggregating it into in-memory buffers
until the configured count or size of the aggregate meets
a threshold and new data is produced. This supports multiple
data aggregation patterns:
	- concatenate batches of data with a separator value
	- store batches of data in a JSON array
	- organize nested JSON in a JSON array based on unique keys

The processor supports these patterns:
	JSON array:
		foo bar baz qux >>> {"aggregate":["foo","bar","baz","qux"]}
		{"foo":"bar"} {"baz":"qux"} >>> {"aggregate":[{"foo":"bar"},{"baz":"qux"}]}
	data:
		foo bar baz qux >>> foo\nbar\nbaz\qux
		{"foo":"bar"} {"baz":"qux"} >>> {"foo":"bar"}\n{"baz":"qux"}

The processor uses this Jsonnet configuration:
	{
		type: 'aggregate',
		settings: {
			options: {
				max_count: 1000,
				max_size: 1000,
			},
			output_key: 'aggregate.-1',
		},
	}
*/
type Aggregate struct {
	Options   AggregateOptions         `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Aggregate processor. Conditions are optionally applied on the bytes to enable processing.
func (p Aggregate) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
	}

	// aggregateKeys is used to return elements stored in the
	// buffer in order if the aggregate doesn't meet the
	// configured threshold. any aggregate that meets the
	// threshold is delivered immediately, out of order.
	var aggregateKeys []string
	buffer := map[string]*aggregate.Bytes{}

	if p.Options.MaxCount == 0 {
		p.Options.MaxCount = 1000
	}

	if p.Options.MaxSize == 0 {
		p.Options.MaxSize = 10000
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

		// data that exceeds the size of the buffer will never
		// fit within it
		if len(data) > p.Options.MaxSize {
			return nil, fmt.Errorf("aggregate: size limit %d reached (%d): %w", p.Options.MaxSize, len(data), AggregateBufferSizeLimit)
		}

		var aggregateKey string
		if p.Options.AggregateKey != "" {
			aggregateKey = json.Get(data, p.Options.AggregateKey).String()
		}

		if _, ok := buffer[aggregateKey]; !ok {
			buffer[aggregateKey] = &aggregate.Bytes{}
			buffer[aggregateKey].New(p.Options.MaxSize, p.Options.MaxCount)
			aggregateKeys = append(aggregateKeys, aggregateKey)
		}

		ok, err = buffer[aggregateKey].Add(data)
		if err != nil {
			return nil, fmt.Errorf("aggregate: %v", err)
		}

		// data was successfully added to the buffer, every item after
		// this is a failure
		if ok {
			continue
		}

		elements := buffer[aggregateKey].Get()
		if p.OutputKey != "" {
			var tmp []byte
			for _, element := range elements {
				var err error

				tmp, err = json.Set(tmp, p.OutputKey, element)
				if err != nil {
					return nil, fmt.Errorf("aggregate: %v", err)
				}
			}

			slice = append(slice, tmp)
		} else {
			tmp := bytes.Join(elements, []byte(p.Options.Separator))
			slice = append(slice, tmp)
		}

		// by this point, addition of the failed data is guaranteed to
		// succeed after the buffer is reset
		buffer[aggregateKey].Reset()
		buffer[aggregateKey].Add(data)
	}

	// remaining items must be drained from the buffer, otherwise data is lost
	for _, key := range aggregateKeys {
		if buffer[key].Count() == 0 {
			continue
		}

		elements := buffer[key].Get()
		if p.OutputKey != "" {
			var tmp []byte
			for _, element := range elements {
				var err error

				tmp, err = json.Set(tmp, p.OutputKey, element)
				if err != nil {
					return nil, fmt.Errorf("aggregate: %v", err)
				}
			}

			slice = append(slice, tmp)
		} else {
			tmp := bytes.Join(elements, []byte(p.Options.Separator))
			slice = append(slice, tmp)
		}
	}

	return slice, nil
}
