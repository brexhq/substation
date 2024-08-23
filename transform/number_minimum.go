package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/message"
)

func newNumberMinimum(_ context.Context, cfg config.Config) (*numberMinimum, error) {
	conf := numberValConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform number_minimum: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "number_minimum"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := numberMinimum{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type numberMinimum struct {
	conf     numberValConfig
	isObject bool
}

func (tf *numberMinimum) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

	flo64 := math.Min(value.Float(), tf.conf.Value)

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

func (tf *numberMinimum) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
