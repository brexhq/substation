package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/internal/aggregate"
	"github.com/brexhq/substation/v2/message"
)

func newAggregateToString(_ context.Context, cfg config.Config) (*aggregateToString, error) {
	conf := aggregateStrConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform aggregate_to_string: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "aggregate_to_string"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := aggregateToString{
		conf:      conf,
		separator: []byte(conf.Separator),
	}

	agg, err := aggregate.New(aggregate.Config{
		Count:    conf.Batch.Count,
		Size:     conf.Batch.Size,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}
	tf.agg = agg

	return &tf, nil
}

type aggregateToString struct {
	conf aggregateStrConfig

	separator []byte

	mu  sync.Mutex
	agg *aggregate.Aggregate
}

func (tf *aggregateToString) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		var output []*message.Message

		for _, items := range tf.agg.GetAll() {
			agg := aggToStr(items.Get(), tf.separator)
			outMsg := message.New().SetData(agg)

			output = append(output, outMsg)
		}

		tf.agg.ResetAll()

		output = append(output, msg)
		return output, nil
	}

	// If this value does not exist, then all data is batched together.
	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if ok := tf.agg.Add(key, msg.Data()); ok {
		return nil, nil
	}

	agg := aggToStr(tf.agg.Get(key), tf.separator)
	outMsg := message.New().SetData(agg)

	// If data cannot be added after reset, then the batch is misconfgured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errBatchNoMoreData)
	}

	return []*message.Message{outMsg}, nil
}

func (tf *aggregateToString) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
