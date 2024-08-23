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

	insp := metaAny{
		conf: conf,
	}

	insp.inspectors = make([]Inspector, len(conf.Inspectors))
	for i, c := range conf.Inspectors {
		cond, err := New(ctx, c)
		if err != nil {
			return nil, err
		}

		insp.inspectors[i] = cond
	}

	return &insp, nil
}

type metaAny struct {
	conf metaConfig

	inspectors []Inspector
}

func (c *metaAny) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if c.conf.Object.SourceKey == "" {
		for _, cnd := range c.inspectors {
			ok, err := cnd.Inspect(ctx, msg)
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
		for _, cnd := range c.inspectors {
			ok, err := cnd.Inspect(ctx, m)
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
		for _, cnd := range c.inspectors {
			ok, err := cnd.Inspect(ctx, m)
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
