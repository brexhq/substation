package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/message"
)

func newAggregateToArray(_ context.Context, cfg config.Config) (*aggregateToArray, error) {
	conf := aggregateArrayConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform aggregate_to_array: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "aggregate_to_array"
	}

	tf := aggregateToArray{
		conf:      conf,
		hasObjTrg: conf.Object.TargetKey != "",
	}

	agg, err := aggregate.New(aggregate.Config{
		Count:    conf.Batch.Count,
		Size:     conf.Batch.Size,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}
	tf.agg = *agg

	return &tf, nil
}

type aggregateToArray struct {
	conf      aggregateArrayConfig
	hasObjTrg bool

	mu  sync.Mutex
	agg aggregate.Aggregate
}

func (tf *aggregateToArray) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		var output []*message.Message

		for _, items := range tf.agg.GetAll() {
			array := aggToArray(items.Get())

			outMsg := message.New()
			if tf.hasObjTrg {
				if err := outMsg.SetValue(tf.conf.Object.TargetKey, array); err != nil {
					return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
				}
			} else {
				outMsg.SetData(array)
			}

			output = append(output, outMsg)
		}

		tf.agg.ResetAll()

		output = append(output, msg)
		return output, nil
	}

	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if ok := tf.agg.Add(key, msg.Data()); ok {
		return nil, nil
	}

	array := aggToArray(tf.agg.Get(key))

	outMsg := message.New()
	if tf.hasObjTrg {
		if err := outMsg.SetValue(tf.conf.Object.TargetKey, array); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	} else {
		outMsg.SetData(array)
	}

	// If data cannot be added after reset, then the batch is misconfgured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendBatchMisconfigured)
	}

	return []*message.Message{outMsg}, nil
}

func (tf *aggregateToArray) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
