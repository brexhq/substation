package condition

import (
	"context"

	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/v2/config"
)

func newMetaAll(ctx context.Context, cfg config.Config) (*metaAll, error) {
	conf := metaConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	cnd := metaAll{
		conf: conf,
	}

	cnd.inspectors = make([]Inspector, len(conf.Inspectors))
	for i, c := range conf.Inspectors {
		cond, err := New(ctx, c)
		if err != nil {
			return nil, err
		}

		cnd.inspectors[i] = cond
	}

	return &cnd, nil
}

type metaAll struct {
	conf metaConfig

	inspectors []Inspector
}

func (c *metaAll) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if c.conf.Object.SourceKey == "" {
		for _, cnd := range c.inspectors {
			ok, err := cnd.Inspect(ctx, msg)
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
		for _, cnd := range c.inspectors {
			ok, err := cnd.Inspect(ctx, m)
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
		for _, cnd := range c.inspectors {
			ok, err := cnd.Inspect(ctx, m)
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
