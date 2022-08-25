package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
Flatten processes data by flattening JSON arrays. The processor supports these patterns:
	JSON:
		{"flatten":["foo",["bar"]]} >>> {"flatten":["foo","bar"]}

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "flatten",
		"settings": {
			"input_key": "flatten",
			"output_key": "flatten"
		}
	}
*/
type Flatten struct {
	Options   FlattenOptions   `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
FlattenOptions contains custom options settings for the Flatten processor:
	Deep (optional):
		deeply flattens nested arrays
*/
type FlattenOptions struct {
	Deep bool `json:"deep"`
}

// ApplyBatch processes a slice of encapsulated data with the Flatten processor. Conditions are optionally applied to the data to enable processing.
func (p Flatten) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process flatten applybatch: %v", err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("process flatten applybatch: %v", err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Flatten processor.
func (p Flatten) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return cap, fmt.Errorf("process flatten apply: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, ProcessorInvalidDataPattern)
	}

	var value interface{}
	if p.Options.Deep {
		value = cap.Get(p.InputKey + `|@flatten:{"deep":true}`)
	} else {
		value = cap.Get(p.InputKey + `|@flatten`)
	}

	if err := cap.Set(p.OutputKey, value); err != nil {
		return cap, fmt.Errorf("process flatten apply: %v", err)
	}

	return cap, nil
}
