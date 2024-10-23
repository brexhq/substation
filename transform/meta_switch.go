package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/condition"
	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type metaSwitchCaseConfig struct {
	// Condition that must be true for the transforms to be applied.
	Condition config.Config `json:"condition"`
	// Transforms that are applied in series when the condition is true.
	Transforms []config.Config `json:"transforms"`
}

type metaSwitchConfig struct {
	// Cases are the transforms that are conditionally applied. If
	// no condition is configured, then the transform is always
	// applied.
	Cases []metaSwitchCaseConfig `json:"cases"`

	ID string `json:"id"`
}

func (c *metaSwitchConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaSwitchConfig) Validate() error {
	if len(c.Cases) == 0 {
		return fmt.Errorf("cases: %v", iconfig.ErrMissingRequiredOption)
	}

	for _, c := range c.Cases {
		if len(c.Transforms) == 0 {
			return fmt.Errorf("transform: %v", iconfig.ErrMissingRequiredOption)
		}

		for _, t := range c.Transforms {
			if t.Type == "" {
				return fmt.Errorf("type: %v", iconfig.ErrMissingRequiredOption)
			}
		}
	}

	return nil
}

type metaSwitchConditional struct {
	condition    condition.Conditioner
	transformers []Transformer
}

type metaSwitchDefaultInspector struct{}

func (i *metaSwitchDefaultInspector) Condition(ctx context.Context, msg *message.Message) (bool, error) {
	return true, nil
}

func newMetaSwitch(ctx context.Context, cfg config.Config) (*metaSwitch, error) {
	conf := metaSwitchConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform meta_switch: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "meta_switch"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	conditionals := make([]metaSwitchConditional, len(conf.Cases))
	for i, s := range conf.Cases {
		conditional := metaSwitchConditional{}

		// If no condition is configured, then the transforms are always
		// applied.
		conditional.condition = &metaSwitchDefaultInspector{}
		if s.Condition.Type != "" {
			op, err := condition.New(ctx, s.Condition)
			if err != nil {
				return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
			}
			conditional.condition = op
		}

		for _, c := range s.Transforms {
			tf, err := New(ctx, c)
			if err != nil {
				return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
			}

			conditional.transformers = append(conditional.transformers, tf)
		}

		conditionals[i] = conditional
	}

	tf := metaSwitch{
		conf:        conf,
		conditional: conditionals,
	}

	return &tf, nil
}

type metaSwitch struct {
	conf metaSwitchConfig

	conditional []metaSwitchConditional
}

func (tf *metaSwitch) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		var messages []*message.Message
		for _, c := range tf.conditional {
			msgs, err := Apply(ctx, c.transformers, msg)
			if err != nil {
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}

			messages = append(messages, msgs...)
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
		ok, err := c.condition.Condition(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		if !ok {
			continue
		}

		msgs, err := Apply(ctx, c.transformers, msg)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
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
