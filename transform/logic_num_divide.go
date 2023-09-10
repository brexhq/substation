package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

func newLogicNumDivide(_ context.Context, cfg config.Config) (*logicNumDivide, error) {
	conf := logicNumConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_math: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_mod_math: %v", err)
	}

	tf := logicNumDivide{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type logicNumDivide struct {
	conf     logicNumConfig
	isObject bool
}

func (tf *logicNumDivide) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

		vFloat64 /= val.Float()
	}

	if !tf.isObject {
		b := []byte(fmt.Sprintf("%v", vFloat64))
		msg.SetData(b)

		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, vFloat64); err != nil {
		return nil, fmt.Errorf("transform: mod_math: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *logicNumDivide) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*logicNumDivide) Close(context.Context) error {
	return nil
}