package condition

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
	"golang.org/x/exp/slices"

	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

type metaForEachConfig struct {
	Object iconfig.Object `json:"object"`

	// Type determines the method of combining results from the inspector.
	//
	// Must be one of:
	//
	// - none: none of the elements match the condition
	//
	// - any: at least one of the elements match the condition
	//
	// - all: all of the elements match the condition
	Type string `json:"type"`
	// Inspector is the condition applied to each element.
	Inspector config.Config `json:"inspector"`
}

func (c *metaForEachConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaForEachConfig) Validate() error {
	if c.Type == "" {
		return fmt.Errorf("type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"none",
			"any",
			"all",
		},
		c.Type) {
		return fmt.Errorf("type %q: %v", c.Type, errors.ErrInvalidOption)
	}

	if c.Inspector.Type == "" {
		return fmt.Errorf("inspector: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaForEach(ctx context.Context, cfg config.Config) (*metaForEach, error) {
	conf := metaForEachConfig{}
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

	meta := metaForEach{
		conf: conf,
		insp: i,
	}

	return &meta, nil
}

type metaForEach struct {
	conf metaForEachConfig

	insp inspector
}

func (c *metaForEach) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	var results []bool
	for _, res := range msg.GetValue(c.conf.Object.Key).Array() {
		data := []byte(res.String())
		msg := message.New().SetData(data)

		inspected, err := c.insp.Inspect(ctx, msg)
		if err != nil {
			return false, fmt.Errorf("condition: meta_for_each: %w", err)
		}
		results = append(results, inspected)
	}

	total := len(results)
	matched := 0
	for _, v := range results {
		if v {
			matched++
		}
	}

	switch c.conf.Type {
	case "any":
		return matched > 0, nil
	case "all":
		return total == matched, nil
	case "none":
		return matched == 0, nil
	}

	return false, nil
}

func (c *metaForEach) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
