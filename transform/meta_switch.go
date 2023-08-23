package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type metaSwitchConfig struct {
	Switch []struct {
		Condition condition.Config `json:"condition"`
		Transform config.Config    `json:"transform"`
	} `json:"switch"`
}

type metaSwitch struct {
	conf metaSwitchConfig

	conditional []struct {
		op condition.Operator
		tf Transformer
	}
}

func newMetaSwitch(ctx context.Context, cfg config.Config) (*metaSwitch, error) {
	conf := metaSwitchConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if len(conf.Switch) == 0 {
		return nil, fmt.Errorf("transform: meta_switch: switch: %v", errors.ErrMissingRequiredOption)
	}

	var conditional []struct {
		op condition.Operator
		tf Transformer
	}
	for _, s := range conf.Switch {
		op, err := condition.New(ctx, s.Condition)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_switch: %v", err)
		}

		tf, err := New(ctx, s.Transform)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_switch: %v", err)
		}

		conditional = append(conditional, struct {
			op condition.Operator
			tf Transformer
		}{
			op: op,
			tf: tf,
		})
	}

	meta := metaSwitch{
		conf:        conf,
		conditional: conditional,
	}

	return &meta, nil
}

func (meta *metaSwitch) String() string {
	b, _ := gojson.Marshal(meta.conf)
	return string(b)
}

func (*metaSwitch) Close(context.Context) error {
	return nil
}

func (meta *metaSwitch) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	for _, c := range meta.conditional {
		ok, err := c.op.Operate(ctx, message)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_switch: %v", err)
		}

		if !ok {
			continue
		}

		msgs, err := c.tf.Transform(ctx, message)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_switch: %v", err)
		}

		return msgs, nil
	}

	// If no conditions match, then return the original message.
	return []*mess.Message{message}, nil
}
