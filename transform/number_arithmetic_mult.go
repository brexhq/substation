package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

func newNumberArithmeticMult(_ context.Context, cfg config.Config) (*numberArithmeticMult, error) {
	conf := numberArithmeticConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_number_arithmetic_multiply: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_number_arithmetic_multiply: %v", err)
	}

	tf := numberArithmeticMult{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type numberArithmeticMult struct {
	conf     numberArithmeticConfig
	isObject bool
}

func (tf *numberArithmeticMult) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var value message.Value
	if tf.isObject {
		value = msg.GetValue(tf.conf.Object.Key)
	} else {
		value = bytesToValue(msg.Data())
	}

	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if !value.IsArray() {
		return []*message.Message{msg}, nil
	}

	if len(value.Array()) <= 1 {
		return []*message.Message{msg}, nil
	}
	var vFloat64 float64
	for i, val := range value.Array() {
		if i == 0 {
			vFloat64 = val.Float()
			continue
		}

		vFloat64 *= val.Float()
	}

	if !tf.isObject {
		b := []byte(fmt.Sprintf("%v", vFloat64))
		msg.SetData(b)

		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, vFloat64); err != nil {
		return nil, fmt.Errorf("transform: arithmetic_multiply: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *numberArithmeticMult) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
