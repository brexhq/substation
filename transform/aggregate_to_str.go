package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/message"
)

func newAggregateToStr(_ context.Context, cfg config.Config) (*aggregateToStr, error) {
	conf := aggregateStrConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_aggregate_to_str: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_aggregate_to_str: %v", err)
	}

	tf := aggregateToStr{
		conf:      conf,
		separator: []byte(conf.Separator),
	}

	buffer, err := aggregate.New(aggregate.Config{
		Count:    conf.Buffer.Count,
		Size:     conf.Buffer.Size,
		Interval: conf.Buffer.Interval,
	})
	if err != nil {
		return nil, fmt.Errorf("transform: new_aggregate_to_str: %v", err)
	}

	tf.buffer = buffer
	tf.bufferKey = conf.Buffer.Key

	return &tf, nil
}

type aggregateToStr struct {
	conf aggregateStrConfig

	separator []byte

	// buffer is safe for concurrent access.
	buffer    *aggregate.Aggregate
	bufferKey string
}

func (tf *aggregateToStr) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		var output []*message.Message

		for _, items := range tf.buffer.GetAll() {
			agg := aggToStr(items.Get(), tf.separator)
			outMsg := message.New().SetData(agg)

			output = append(output, outMsg)
		}

		tf.buffer.ResetAll()

		output = append(output, msg)
		return output, nil
	}

	value := msg.GetValue(tf.bufferKey)
	key := value.String()
	if ok := tf.buffer.Add(key, msg.Data()); ok {
		return nil, nil
	}

	agg := aggToStr(tf.buffer.Get(key), tf.separator)
	outMsg := message.New().SetData(agg)

	// By this point, addition of the failed data is guaranteed
	// to succeed after the buffer is reset.
	tf.buffer.Reset(key)
	_ = tf.buffer.Add(key, msg.Data())

	return []*message.Message{outMsg}, nil
}

func (tf *aggregateToStr) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*aggregateToStr) Close(context.Context) error {
	return nil
}
