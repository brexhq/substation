package process

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// forEach processes data by iterating and applying a processor to each element
// in an object array. If multiple processors need to be applied to each element,
// then the pipeline processor should be used to create a nested data processing
// workflow.
//
// This processor supports the object handling pattern.
type procForEach struct {
	process
	Options procForEachOptions `json:"options"`

	procCfg config.Config
	applier Applier
}

type procForEachOptions struct {
	// Processor applied to each element in the object array.
	Processor config.Config `json:"processor"`
}

// Create a new "for each" processor.
func newProcForEach(ctx context.Context, cfg config.Config) (p procForEach, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procForEach{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procForEach{}, err
	}

	// only supports JSON arrays, fail if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return procForEach{}, fmt.Errorf("process: for_each: options %+v: %v", p.Options, errors.ErrMissingRequiredOption)
	}

	// configured processor is converted to a JSON object so that the
	// settings can be modified and put into a new processor
	// we cannot directly modify p.Options.Processor, otherwise we will
	// cause errors during iteration
	conf, _ := gojson.Marshal(p.Options.Processor)

	inputKey := p.Options.Processor.Type
	if innerKey, ok := p.Options.Processor.Settings["key"].(string); ok && innerKey != "" {
		inputKey = p.Options.Processor.Type + "." + innerKey
	}
	conf, _ = json.Set(conf, "settings.key", inputKey)

	outputKey := p.Options.Processor.Type
	if innerKey, ok := p.Options.Processor.Settings["set_key"].(string); ok && innerKey != "" {
		outputKey = p.Options.Processor.Type + "." + innerKey
	}
	conf, _ = json.Set(conf, "settings.set_key", outputKey)

	if err := gojson.Unmarshal(conf, &p.procCfg); err != nil {
		return procForEach{}, err
	}

	p.applier, err = NewApplier(ctx, p.procCfg)
	if err != nil {
		return procForEach{}, fmt.Errorf("process: for_each: %v", err)
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procForEach) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procForEach) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procForEach) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procForEach) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	result := capsule.Get(p.Key)
	if !result.IsArray() {
		return capsule, nil
	}

	for _, res := range result.Array() {
		tmpCap := config.NewCapsule()
		if err := tmpCap.Set(p.procCfg.Type, res); err != nil {
			return capsule, fmt.Errorf("process: for_each: %v", err)
		}

		tmpCap, err := p.applier.Apply(ctx, tmpCap)
		if err != nil {
			return capsule, fmt.Errorf("process: for_each: %v", err)
		}

		value := tmpCap.Get(p.procCfg.Type)
		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process: for_each: %v", err)
		}
	}

	return capsule, nil
}
