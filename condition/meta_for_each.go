package condition

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
	"golang.org/x/exp/slices"

	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

type metaInspForEachConf struct {
	// Key is the message key used during inspection.
	Key string `json:"key"`
	// Negate is a boolean that negates the inspection result.
	Negate bool `json:"negate"`
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

type metaInspForEach struct {
	conf metaInspForEachConf

	inspector Inspector
}

func newMetaInspForEach(ctx context.Context, cfg config.Config) (*metaInspForEach, error) {
	conf := metaInspForEachConf{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Type == "" {
		return nil, fmt.Errorf("condition: meta_for_each: type: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Inspector.Type == "" {
		return nil, fmt.Errorf("condition: meta_for_each: inspector: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"none",
			"any",
			"all",
		},
		conf.Type) {
		return nil, fmt.Errorf("condition: meta_for_each: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	i, err := NewInspector(ctx, conf.Inspector)
	if err != nil {
		return nil, fmt.Errorf("condition: meta_for_each: %v", err)
	}

	meta := metaInspForEach{
		conf:      conf,
		inspector: i,
	}

	return &meta, nil
}

func (c *metaInspForEach) String() string {
	b, _ := gojson.Marshal(c.conf)
	return string(b)
}

func (c *metaInspForEach) Inspect(ctx context.Context, message *mess.Message) (output bool, err error) {
	if message.IsControl() {
		return false, nil
	}

	var results []bool
	for _, res := range message.Get(c.conf.Key).Array() {
		tmpCapule, err := mess.New(
			mess.SetData([]byte(res.String())),
		)
		if err != nil {
			return false, fmt.Errorf("condition: meta_for_each: %w", err)
		}

		inspected, err := c.inspector.Inspect(ctx, tmpCapule)
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
		output = matched > 0
	case "all":
		output = total == matched
	case "none":
		output = matched == 0
	}

	if c.conf.Negate {
		return !output, nil
	}

	return output, nil
}
