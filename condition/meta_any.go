package condition

import (
	"context"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newMetaAny(ctx context.Context, cfg config.Config) (*metaAny, error) {
	conf := metaConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	cnd := metaAny{
		conf: conf,
	}

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

type metaAny struct {
	conf metaConfig

	cnds []Conditioner
}

func (c *metaAny) Condition(ctx context.Context, msg *message.Message) (bool, error) {
	if c.conf.Object.SourceKey == "" {
		for _, cnd := range c.cnds {
			ok, err := cnd.Condition(ctx, msg)
			if err != nil {
				return false, err
			}

			if ok {
				return true, nil
			}
		}

		return false, nil
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

			if ok {
				return true, nil
			}
		}

		return false, nil
	}

	for _, v := range value.Array() {
		m := message.New().SetData(v.Bytes()).SetMetadata(msg.Metadata())
		for _, cnd := range c.cnds {
			ok, err := cnd.Condition(ctx, m)
			if err != nil {
				return false, err
			}

			// If any of the values in the array match, then this returns true.
			if ok {
				return true, nil
			}
		}
	}

	return false, nil
}
