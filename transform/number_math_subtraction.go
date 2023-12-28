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

func newNumberMathSubtraction(_ context.Context, cfg config.Config) (*numberMathSubtraction, error) {
	conf := numberMathConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: number_math_subtraction: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: number_math_subtraction: %v", err)
	}

	tf := numberMathSubtraction{
		conf:     conf,
		isObject: conf.Object.SrcKey != "" && conf.Object.DstKey != "",
	}

	return &tf, nil
}

type numberMathSubtraction struct {
	conf     numberMathConfig
	isObject bool
}

func (tf *numberMathSubtraction) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var value message.Value
	if tf.isObject {
		value = msg.GetValue(tf.conf.Object.SrcKey)
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

		vFloat64 -= val.Float()
	}

	strFloat64 := numberFloat64ToString(vFloat64)
	if !tf.isObject {
		msg.SetData([]byte(strFloat64))

		return []*message.Message{msg}, nil
	}

	f, err := strconv.ParseFloat(strFloat64, 64)
	if err != nil {
		return nil, fmt.Errorf("transform: number_math_subtraction: %v", err)
	}

	if err := msg.SetValue(tf.conf.Object.DstKey, f); err != nil {
		return nil, fmt.Errorf("transform: number_math_subtraction: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *numberMathSubtraction) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
