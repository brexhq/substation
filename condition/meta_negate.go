package condition

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"

	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

type metaNegateConfig struct {
	Inspector config.Config `json:"inspector"`
}

func (c *metaNegateConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaNegateConfig) Validate() error {
	if c.Inspector.Type == "" {
		return fmt.Errorf("inspector: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaNegate(ctx context.Context, cfg config.Config) (*metaNegate, error) {
	conf := metaNegateConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	i, err := newInspector(ctx, conf.Inspector)
	if err != nil {
		return nil, fmt.Errorf("condition: meta_for_each: %v", err)
	}

	meta := metaNegate{
		conf: conf,
		insp: i,
	}

	return &meta, nil
}

type metaNegate struct {
	conf metaNegateConfig

	insp inspector
}

func (c *metaNegate) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	match, err := c.insp.Inspect(ctx, msg)
	if err != nil {
		return false, err
	}

	return !match, nil
}

func (c *metaNegate) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
