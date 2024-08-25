package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

func newNumberMaximum(_ context.Context, cfg config.Config) (*numberMaximum, error) {
	conf := numberValConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform number_maximum: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "number_maximum"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := numberMaximum{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type numberMaximum struct {
	conf     numberValConfig
	isObject bool
}

func (tf *numberMaximum) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var value message.Value
	if tf.isObject {
		value = msg.GetValue(tf.conf.Object.SourceKey)
	} else {
		value = bytesToValue(msg.Data())
	}

	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	flo64 := math.Max(value.Float(), tf.conf.Value)

	if !tf.isObject {
		s := numberFloat64ToString(flo64)
		msg.SetData([]byte(s))

		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, flo64); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *numberMaximum) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
