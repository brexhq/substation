package process

import (
	"bytes"
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
	"github.com/jshlbrd/go-aggregate"
)

// errAggregateSizeLimit is returned when the aggregate's buffer size
// limit is reached. If this error occurs, then increase the size of
// the buffer or use the Drop processor to remove data that exceeds
// the buffer limit.
var errAggregateSizeLimit = fmt.Errorf("data exceeded size limit")

// aggregate processes data by buffering and aggregating it into a
// single item.
//
// Multiple data aggregation patterns are supported, including:
//
// - aggregate data using a separator value
//
// - aggregate data into an object array
//
// - aggregate nested objects into object arrays based on unique keys
//
// This processor supports the data and object handling patterns.
type procAggregate struct {
	process
	Options procAggregateOptions `json:"options"`
}

type procAggregateOptions struct {
	// Key retrieves a value from an object that is used to organize
	// aggregated objects.
	//
	// This is only used when handling objects and defaults to an
	// empty string.
	Key string `json:"key"`
	// Separator is the string that joins aggregated data.
	//
	// This is only used when handling data and defaults to an empty
	// string.
	Separator string `json:"separator"`
	// MaxCount determines the maximum number of items stored in the
	// buffer before emitting aggregated data.
	//
	// This is optional and defaults to 1000 items.
	MaxCount int `json:"max_count"`
	// MaxSize determines the maximum size (in bytes) of items stored
	// in the buffer before emitting aggregated data.
	//
	// This is optional and defaults to 10000 (10KB).
	MaxSize int `json:"max_size"`
}

// String returns the processor settings as an object.
func (p procAggregate) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procAggregate) Close(context.Context) error {
	return nil
}

// Create a new aggregate processor.
func newProcAggregate(ctx context.Context, cfg config.Config) (p procAggregate, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procAggregate{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procAggregate{}, err
	}

	if p.Options.MaxCount == 0 {
		p.Options.MaxCount = 1000
	}

	if p.Options.MaxSize == 0 {
		p.Options.MaxSize = 10000
	}

	return p, nil
}

// Stream processes a pipeline of capsules with the processor.
func (p procAggregate) Stream(ctx context.Context, in, out *config.Channel) error {
	defer out.Close()

	// aggregateKeys is used to return data stored in the
	// buffer in order if the aggregate doesn't meet the
	// configured threshold. any aggregate that meets the
	// threshold is delivered immediately, out of order.
	var aggregateKeys []string
	buffer := map[string]*aggregate.Bytes{}

	for capsule := range in.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if ok, err := p.operator.Operate(ctx, capsule); err != nil {
				return fmt.Errorf("process: aggregate: %v", err)
			} else if !ok {
				out.C <- capsule
				continue
			}

			// data that exceeds the size of the buffer will never
			// fit within it
			length := len(capsule.Data())
			if length > p.Options.MaxSize {
				return fmt.Errorf("process: aggregate: size %d data length %d: %v", p.Options.MaxSize, length, errAggregateSizeLimit)
			}

			var aggregateKey string
			if p.Options.Key != "" {
				aggregateKey = capsule.Get(p.Options.Key).String()
			}

			if _, ok := buffer[aggregateKey]; !ok {
				buffer[aggregateKey] = &aggregate.Bytes{}
				buffer[aggregateKey].New(p.Options.MaxCount, p.Options.MaxSize)
				aggregateKeys = append(aggregateKeys, aggregateKey)
			}

			ok := buffer[aggregateKey].Add(capsule.Data())
			// data was successfully added to the buffer, every item after
			// this is a failure
			if ok {
				continue
			}

			// the buffer is full, emit the aggregated data
			data := buffer[aggregateKey].Get()
			c, err := p.newCapsule(data)
			if err != nil {
				return fmt.Errorf("process: aggregate: %v", err)
			}
			out.Send(c)

			// by this point, addition of the failed data is guaranteed
			// to succeed after the buffer is reset
			buffer[aggregateKey].Reset()
			_ = buffer[aggregateKey].Add(capsule.Data())
		}
	}

	// emit any remaining data in the buffer
	for _, aggregateKey := range aggregateKeys {
		data := buffer[aggregateKey].Get()
		c, err := p.newCapsule(data)
		if err != nil {
			return fmt.Errorf("process: aggregate: %v", err)
		}
		out.Send(c)
	}

	return nil
}

// Batch processes one or more capsules with the processor.
func (p procAggregate) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	// aggregateKeys is used to return elements stored in the
	// buffer in order if the aggregate doesn't meet the
	// configured threshold. any aggregate that meets the
	// threshold is delivered immediately, out of order.
	var aggregateKeys []string
	buffer := map[string]*aggregate.Bytes{}

	newCapsules := newBatch(&capsules)
	for _, capsule := range capsules {
		if ok, err := p.operator.Operate(ctx, capsule); err != nil {
			return nil, fmt.Errorf("process: aggregate: %v", err)
		} else if !ok {
			newCapsules = append(newCapsules, capsule)
			continue
		}

		// data that exceeds the size of the buffer will never
		// fit within it
		length := len(capsule.Data())
		if length > p.Options.MaxSize {
			return nil, fmt.Errorf("process: aggregate: size %d data length %d: %v", p.Options.MaxSize, length, errAggregateSizeLimit)
		}

		var aggregateKey string
		if p.Options.Key != "" {
			aggregateKey = capsule.Get(p.Options.Key).String()
		}

		if _, ok := buffer[aggregateKey]; !ok {
			buffer[aggregateKey] = &aggregate.Bytes{}
			buffer[aggregateKey].New(p.Options.MaxCount, p.Options.MaxSize)
			aggregateKeys = append(aggregateKeys, aggregateKey)
		}

		ok := buffer[aggregateKey].Add(capsule.Data())
		// data was successfully added to the buffer, every item after
		// this is a failure
		if ok {
			continue
		}

		data := buffer[aggregateKey].Get()
		c, err := p.newCapsule(data)
		if err != nil {
			return nil, fmt.Errorf("process: aggregate: %v", err)
		}
		newCapsules = append(newCapsules, c)

		// by this point, addition of the failed data is guaranteed
		// to succeed after the buffer is reset
		buffer[aggregateKey].Reset()
		_ = buffer[aggregateKey].Add(capsule.Data())
	}

	// remaining items must be drained from the buffer, otherwise
	// data is lost
	for _, key := range aggregateKeys {
		if buffer[key].Count() == 0 {
			continue
		}

		data := buffer[key].Get()
		c, err := p.newCapsule(data)
		if err != nil {
			return nil, fmt.Errorf("process: aggregate: %v", err)
		}

		newCapsules = append(newCapsules, c)
	}

	return newCapsules, nil
}

func (p procAggregate) newCapsule(data [][]byte) (config.Capsule, error) {
	newCapsule := config.NewCapsule()

	if p.SetKey != "" {
		var value []byte
		for _, element := range data {
			var err error

			value, err = json.Set(value, p.SetKey, element)
			if err != nil {
				return newCapsule, err
			}
		}

		newCapsule.SetData(value)
		return newCapsule, nil
	}

	value := bytes.Join(data, []byte(p.Options.Separator))
	newCapsule.SetData(value)

	return newCapsule, nil
}
