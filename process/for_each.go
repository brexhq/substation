package process

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
)

/*
ForEach processes data by iterating and applying a processor to each element in a JSON array. The processor supports these patterns:
	JSON:
		{"input":["ABC","DEF"]} >>> {"input":["ABC","DEF"],"output":["abc","def"]}

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "for_each",
		"settings": {
			"options": {
				"processor": {
					"type": "case",
					"settings": {
						"options": {
							"case": "lower"
						}
					}
				}
			},
			input_key: "input",
			output_key: "output.-1"
		}
	}
*/
type ForEach struct {
	Options   ForEachOptions   `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
ForEachOptions contains custom options for the ForEach processor:
	Processor:
		processor to apply to the data
*/
type ForEachOptions struct {
	Processor config.Config
}

// ApplyBatch processes a slice of encapsulated data with the ForEach processor. Conditions are optionally applied to the data to enable processing.
func (p ForEach) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process for_each applybatch: %v", err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("process for_each applybatch: %v", err)
	}

	return caps, nil
}

/*
Apply processes encapsulated data with the ForEeach processor.

JSON values are treated as arrays and the configured
processor is applied to each element in the array. If multiple
processors need to be applied to each element, then the
Pipeline processor should be used to create a nested data
processing workflow. For example:
	ForEach -> Pipeline -> [Copy, Delete, Copy]
*/
func (p ForEach) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return cap, fmt.Errorf("process for_each apply: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, ProcessorInvalidDataPattern)
	}

	// configured processor is converted to a JSON object so that the
	// settings can be modified and put into a new processor
	// we cannot directly modify p.Options.Processor, otherwise we will
	// cause errors during iteration
	conf, _ := gojson.Marshal(p.Options.Processor)

	var inputKey, outputKey string
	if _, ok := p.Options.Processor.Settings["input_key"]; ok {
		inputKey = p.Options.Processor.Type + "." + p.Options.Processor.Settings["input_key"].(string)
	} else {
		inputKey = p.Options.Processor.Type
	}
	conf, _ = json.Set(conf, "settings.input_key", inputKey)

	if _, ok := p.Options.Processor.Settings["output_key"]; ok {
		outputKey = p.Options.Processor.Type + "." + p.Options.Processor.Settings["output_key"].(string)
	} else {
		outputKey = p.Options.Processor.Type
	}
	conf, _ = json.Set(conf, "settings.output_key", outputKey)

	var processor config.Config
	gojson.Unmarshal(conf, &processor)

	applicator, err := ApplicatorFactory(processor)
	if err != nil {
		return cap, fmt.Errorf("process for_each apply: %v", err)
	}

	result := cap.Get(p.InputKey)
	if !result.IsArray() {
		return cap, nil
	}

	for _, res := range result.Array() {
		tmpCap := config.NewCapsule()
		if err := tmpCap.Set(processor.Type, res); err != nil {
			return cap, fmt.Errorf("process for_each apply: %v", err)
		}

		tmpCap, err = applicator.Apply(ctx, tmpCap)
		if err != nil {
			return cap, fmt.Errorf("process for_each apply: %v", err)
		}

		value := tmpCap.Get(processor.Type)
		if err := cap.Set(p.OutputKey, value); err != nil {
			return cap, fmt.Errorf("process for_each apply: %v", err)
		}
	}

	return cap, nil
}
