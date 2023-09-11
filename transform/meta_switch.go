package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type metaSwitchConfig struct {
	Switch []struct {
		Condition condition.Config `json:"condition"`
		Transform config.Config    `json:"transform"`
	} `json:"switch"`
}

func (c *metaSwitchConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaSwitchConfig) Validate() error {
	if len(c.Switch) == 0 {
		return fmt.Errorf("switch: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaSwitch(ctx context.Context, cfg config.Config) (*metaSwitch, error) {
	conf := metaSwitchConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_meta_switch: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_meta_switch: %v", err)
	}

	var conditional []struct {
		op condition.Operator
		tf Transformer
	}
	for _, s := range conf.Switch {
		op, err := condition.New(ctx, s.Condition)
		if err != nil {
			return nil, fmt.Errorf("transform: new_meta_switch: %v", err)
		}

		tf, err := New(ctx, s.Transform)
		if err != nil {
			return nil, fmt.Errorf("transform: new_meta_switch: %v", err)
		}

		conditional = append(conditional, struct {
			op condition.Operator
			tf Transformer
		}{
			op: op,
			tf: tf,
		})
	}

	tf := metaSwitch{
		conf:        conf,
		conditional: conditional,
	}

	return &tf, nil
}

type metaSwitch struct {
	conf metaSwitchConfig

	conditional []struct {
		op condition.Operator
		tf Transformer
	}
}

func (meta *metaSwitch) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	for _, c := range meta.conditional {
		ok, err := c.op.Operate(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_switch: %v", err)
		}

		if !ok {
			continue
		}

		msgs, err := c.tf.Transform(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_switch: %v", err)
		}

		return msgs, nil
	}

	// If no conditions match, then return the original message.
	return []*message.Message{msg}, nil
}

func (meta *metaSwitch) String() string {
	b, _ := json.Marshal(meta.conf)
	return string(b)
}

func (*metaSwitch) Close(context.Context) error {
	return nil
}
