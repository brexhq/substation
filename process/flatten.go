package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

type flatten struct {
	process
	Options flattenOptions `json:"options"`
}

type flattenOptions struct {
	Deep bool `json:"deep"`
}

// Close closes resources opened by the flatten processor.
func (p flatten) Close(context.Context) error {
	return nil
}

func (p flatten) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process flatten: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the flatten processor.
func (p flatten) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process flatten: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	var value interface{}
	if p.Options.Deep {
		value = capsule.Get(p.Key + `|@flatten:{"deep":true}`)
	} else {
		value = capsule.Get(p.Key + `|@flatten`)
	}

	if err := capsule.Set(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process flatten: %v", err)
	}

	return capsule, nil
}
