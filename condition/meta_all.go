package condition

import (
	"context"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newMetaAll(ctx context.Context, cfg config.Config) (*metaAll, error) {
	conf := metaConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	cnd := metaAll{
		conf: conf,
	}

	// Generate a list of all of the conditions that need to be met
	cnd.cnds = make([]Conditioner, len(conf.Conditions))
	for i, c := range conf.Conditions {
		cond, err := New(ctx, c)
		if err != nil {
			return nil, err
		}

		cnd.cnds[i] = cond
	}

	return &cnd, nil
}

type metaAll struct {
	conf metaConfig

	cnds []Conditioner
}

func (c *metaAll) Condition(ctx context.Context, msg *message.Message) (bool, error) {
	if c.conf.Object.SourceKey == "" {
		for _, cnd := range c.cnds {
			ok, err := cnd.Condition(ctx, msg)
			if err != nil {
				return false, err
			}

			if !ok {
				return false, nil
			}
		}

		return true, nil
	}

	value := msg.GetValue(c.conf.Object.SourceKey)
	if !value.Exists() {
		return false, nil
	}

	if !value.IsArray() {
		m := message.New().SetData(value.Bytes()).SetMetadata(msg.Metadata())
		for _, cnd := range c.cnds {
			ok, err := cnd.Condition(ctx, m)
			if err != nil {
				return false, err
			}

			if !ok {
				return false, nil
			}
		}

		return true, nil
	}

	for _, v := range value.Array() {
		m := message.New().SetData(v.Bytes()).SetMetadata(msg.Metadata())
		for _, cnd := range c.cnds {
			ok, err := cnd.Condition(ctx, m)
			if err != nil {
				return false, err
			}

			// If any of the values in the array do not match, then this returns false.
			if !ok {
				return false, nil
			}
		}
	}

	// At this point every value in the array matched, so this returns true.
	return true, nil
}
