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
	// Cases are the transforms that are conditionally applied. If
	// no condition is configured, then the transform is always
	// applied.
	Cases []struct {
		Condition condition.Config `json:"condition"`
		Transform config.Config    `json:"transform"`
	} `json:"cases"`
}

func (c *metaSwitchConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaSwitchConfig) Validate() error {
	if len(c.Cases) == 0 {
		return fmt.Errorf("cases: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaSwitch(ctx context.Context, cfg config.Config) (*metaSwitch, error) {
	conf := metaSwitchConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: meta_switch: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: meta_switch: %v", err)
	}

	var conditional []struct {
		op condition.Operator
		tf Transformer
	}
	for _, s := range conf.Cases {
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

func (tf *metaSwitch) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		var messages []*message.Message
		for _, c := range tf.conditional {
			res, err := c.tf.Transform(ctx, msg)
			if err != nil {
				return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
			}

			messages = append(messages, res...)
		}

		// This is required to deduplicate the control messages that
		// were sent to the conditional transforms.
		var msgs []*message.Message
		for _, m := range messages {
			if m.IsControl() {
				continue
			}

			msgs = append(msgs, m)
		}

		msgs = append(msgs, msg)
		return msgs, nil
	}

	for _, c := range tf.conditional {
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

func (tf *metaSwitch) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
