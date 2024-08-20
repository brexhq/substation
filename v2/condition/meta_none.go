package condition

import (
	"context"

	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/v2/config"
)

func newMetaNone(ctx context.Context, cfg config.Config) (*metaNone, error) {
	conf := metaConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := metaNone{
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

type metaNone struct {
	conf metaConfig

	inspectors []Inspector
}

func (c *metaNone) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if c.conf.Object.SourceKey == "" {
		for _, cnd := range c.inspectors {
			ok, err := cnd.Inspect(ctx, msg)
			if err != nil {
				return false, err
			}

			if ok {
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
		m := message.New().SetData(msg.Data()).SetMetadata(msg.Metadata())
		for _, cnd := range c.inspectors {
			ok, err := cnd.Inspect(ctx, m)
			if err != nil {
				return false, err
			}

			if ok {
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

			if ok {
				return false, nil
			}
		}
	}

	return true, nil
}
