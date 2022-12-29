package process

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
)

// forEach processes data by iterating and applying a processor to each element
// in an object array. If multiple processors need to be applied to each element,
// then the pipeline processor should be used to create a nested data processing
// workflow.
//
// This processor supports the object handling pattern.
type _forEach struct {
	process
	Options _forEachOptions `json:"options"`
}

type _forEachOptions struct {
	// Processor applied to each element in the object array.
	Processor config.Config
}

// String returns the processor settings as an object.
func (p _forEach) String() string {
	return toString(p)
}

// Close closes resources opened by the processor.
func (p _forEach) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _forEach) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
func (p _forEach) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process for_each: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
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

	var processor config.Config
	if err := gojson.Unmarshal(conf, &processor); err != nil {
		return capsule, err
	}

	applier, err := ApplierFactory(processor)
	if err != nil {
		return capsule, fmt.Errorf("process for_each: %v", err)
	}

	result := capsule.Get(p.Key)
	if !result.IsArray() {
		return capsule, nil
	}

	for _, res := range result.Array() {
		tmpCap := config.NewCapsule()
		if err := tmpCap.Set(processor.Type, res); err != nil {
			return capsule, fmt.Errorf("process for_each: %v", err)
		}

		tmpCap, err = applier.Apply(ctx, tmpCap)
		if err != nil {
			return capsule, fmt.Errorf("process for_each: %v", err)
		}

		value := tmpCap.Get(processor.Type)
		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process for_each: %v", err)
		}
	}

	return capsule, nil
}
