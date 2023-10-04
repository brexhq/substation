package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/message"
)

func newAggregateToArray(_ context.Context, cfg config.Config) (*aggregateToArray, error) {
	conf := aggregateArrayConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: aggregate_to_array: %v", err)
	}

	tf := aggregateToArray{
		conf:         conf,
		hasObjSetKey: conf.Object.SetKey != "",
	}

	buffer, err := aggregate.New(aggregate.Config{
		Count:    conf.Buffer.Count,
		Size:     conf.Buffer.Size,
		Duration: conf.Buffer.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("transform: aggregate_to_array: %v", err)
	}

	tf.buffer = buffer
	tf.bufferKey = conf.Buffer.Key

	return &tf, nil
}

type aggregateToArray struct {
	conf         aggregateArrayConfig
	hasObjSetKey bool

	// buffer is safe for concurrent access.
	buffer    *aggregate.Aggregate
	bufferKey string
}

func (tf *aggregateToArray) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	//nolint: nestif // ignore nesting complexity
	if msg.IsControl() {
		var output []*message.Message

		for _, items := range tf.buffer.GetAll() {
			agg, err := aggToArray(items.Get())
			if err != nil {
				return nil, fmt.Errorf("transform: aggregate_to_array: %v", err)
			}

			outMsg := message.New()
			if tf.hasObjSetKey {
				if err := outMsg.SetValue(tf.conf.Object.SetKey, agg); err != nil {
					return nil, fmt.Errorf("transform: aggregate_to_array: %v", err)
				}
			} else {
				outMsg.SetData(agg)
			}

			output = append(output, outMsg)
		}

		tf.buffer.ResetAll()

		output = append(output, msg)
		return output, nil
	}

	key := msg.GetValue(tf.bufferKey).String()
	if ok := tf.buffer.Add(key, msg.Data()); ok {
		return nil, nil
	}

	agg, err := aggToArray(tf.buffer.Get(key))
	if err != nil {
		return nil, fmt.Errorf("transform: aggregate_to_array: %v", err)
	}

	outMsg := message.New()
	if tf.hasObjSetKey {
		if err := outMsg.SetValue(tf.conf.Object.SetKey, agg); err != nil {
			return nil, fmt.Errorf("transform: aggregate_to_array: %v", err)
		}
	} else {
		outMsg.SetData(agg)
	}

	// By this point, addition of the failed data is guaranteed
	// to succeed after the buffer is reset.
	tf.buffer.Reset(key)
	_ = tf.buffer.Add(key, msg.Data())

	return []*message.Message{outMsg}, nil
}

func (tf *aggregateToArray) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
