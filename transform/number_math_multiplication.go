package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/message"
)

func newNumberMathMultiplication(_ context.Context, cfg config.Config) (*numberMathMultiplication, error) {
	conf := numberMathConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform number_math_multiplication: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "number_math_multiplication"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := numberMathMultiplication{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type numberMathMultiplication struct {
	conf     numberMathConfig
	isObject bool
}

func (tf *numberMathMultiplication) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

	if !value.IsArray() {
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

	strFloat64 := numberFloat64ToString(vFloat64)
	if !tf.isObject {
		msg.SetData([]byte(strFloat64))

		return []*message.Message{msg}, nil
	}

	f, err := strconv.ParseFloat(strFloat64, 64)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, f); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *numberMathMultiplication) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
