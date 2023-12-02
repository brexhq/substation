package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

func newNumberMathAddition(_ context.Context, cfg config.Config) (*numberMathAddition, error) {
	conf := numberMathConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: number_math_addition: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: number_math_addition: %v", err)
	}

	tf := numberMathAddition{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type numberMathAddition struct {
	conf     numberMathConfig
	isObject bool
}

func (tf *numberMathAddition) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

	var vFloat64 float64
	for i, val := range value.Array() {
		if i == 0 {
			vFloat64 = val.Float()
			continue
		}

		vFloat64 += val.Float()
	}

	strFloat64 := numberFloat64ToString(vFloat64)
	if !tf.isObject {
		msg.SetData([]byte(strFloat64))

		return []*message.Message{msg}, nil
	}

	f, err := strconv.ParseFloat(strFloat64, 64)
	if err != nil {
		return nil, fmt.Errorf("transform: number_math_addition: %v", err)
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, f); err != nil {
		return nil, fmt.Errorf("transform: number_math_addition: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *numberMathAddition) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
