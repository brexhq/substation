package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newAggregateFromString(_ context.Context, cfg config.Config) (*aggregateFromString, error) {
	conf := aggregateStrConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform aggregate_from_string: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "aggregate_from_string"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := aggregateFromString{
		conf:      conf,
		separator: []byte(conf.Separator),
	}

	return &tf, nil
}

type aggregateFromString struct {
	conf aggregateStrConfig

	separator []byte
}

func (tf *aggregateFromString) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
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

func (tf *aggregateFromString) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
