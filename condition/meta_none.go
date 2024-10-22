package condition

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newMetaNone(ctx context.Context, cfg config.Config) (*metaNone, error) {
	conf := metaConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	cnd := metaNone{
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

type metaNone struct {
	conf metaConfig

	cnds []Conditioner
}

func (c *metaNone) Condition(ctx context.Context, msg *message.Message) (bool, error) {
	if c.conf.Object.SourceKey == "" {
		for _, cnd := range c.cnds {
			ok, err := cnd.Condition(ctx, msg)
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
		for _, cnd := range c.cnds {
			ok, err := cnd.Condition(ctx, m)
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
		for _, cnd := range c.cnds {
			ok, err := cnd.Condition(ctx, m)
			if err != nil {
				return false, err
			}

			// If any of the values in the array match, then this returns false.
			if ok {
				fmt.Println("return: false") // Debugging line.
				return false, nil
			}
		}
	}

	// At this point every value in the array did not match, so this returns true.
	return true, nil
}
