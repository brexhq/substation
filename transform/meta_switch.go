package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
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

	logic []struct {
		op    condition.Operator
		tform Transformer
	}
}

func newMetaSwitch(ctx context.Context, cfg config.Config) (*metaSwitch, error) {
	conf := metaSwitchConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if len(conf.Switch) == 0 {
		return nil, fmt.Errorf("transform: meta_switch: switch: %v", errors.ErrMissingRequiredOption)
	}

	var logic []struct {
		op    condition.Operator
		tform Transformer
	}
	for _, s := range conf.Switch {
		op, err := condition.NewOperator(ctx, s.Condition)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_switch: %v", err)
		}

		tform, err := NewTransformer(ctx, s.Transform)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_switch: %v", err)
		}

		logic = append(logic, struct {
			op    condition.Operator
			tform Transformer
		}{
			op:    op,
			tform: tform,
		})
	}

	meta := metaSwitch{
		conf:  conf,
		logic: logic,
	}

	return &meta, nil
}

func (t *metaSwitch) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*metaSwitch) Close(context.Context) error {
	return nil
}

func (t *metaSwitch) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		matched := false
		for _, l := range t.logic {
			ok, err := l.op.Operate(ctx, message)
			if err != nil {
				return nil, fmt.Errorf("transform: meta_switch: %v", err)
			}

			if !ok {
				continue
			}

			matched = true
			messages, err := l.tform.Transform(ctx, message)
			if err != nil {
				return nil, fmt.Errorf("transform: meta_switch: %v", err)
			}

			output = append(output, messages...)

			// If one condition matches, then don't check any more.
			break
		}

		// If no conditions match, then add the message to the output.
		if !matched {
			output = append(output, message)
		}
	}

	return output, nil
}
