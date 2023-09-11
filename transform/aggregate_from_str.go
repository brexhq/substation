package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newAggregateFromStr(_ context.Context, cfg config.Config) (*aggregateFromStr, error) {
	conf := aggregateStrConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_aggregate_from_str: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_aggregate_from_str: %v", err)
	}

	tf := aggregateFromStr{
		conf:      conf,
		separator: []byte(conf.Separator),
	}

	return &tf, nil
}

type aggregateFromStr struct {
	conf aggregateStrConfig

	separator []byte
}

func (tf *aggregateFromStr) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var output []*message.Message
	deagg := aggFromStr(msg.Data(), tf.separator)

	for _, b := range deagg {
		msg := message.New().SetData(b).SetMetadata(msg.Metadata())
		output = append(output, msg)
	}

	return output, nil
}

func (tf *aggregateFromStr) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
