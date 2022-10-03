package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errPipelineArrayInput is returned when the Pipeline processor is configured to process JSON, but the input is an array. Array values are not supported by this processor, instead the input should be run through the ForEach processor (which can encapsulate the Pipeline processor).
const errPipelineArrayInput = errors.Error("input is an array")

/*
Pipeline processes data by applying a series of processors. This processor should be used when data requires complex processing outside of the boundaries of any data structures (see tests for examples). The processor supports these patterns:

	JSON:
		{"pipeline":"H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA"} >>> {"pipeline":"foo"}
	data:
		H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA >> foo

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "pipeline",
		"settings": {
			"options": {
				"processors": [
					{
						"type": "base64",
						"settings": {
							"options": {
								"direction": "from"
							}
						}
					},
					{
						"type": "gzip",
						"settings": {
							"options": {
								"direction": "from"
							}
						}
					}
				]
			},
			"input_key": "pipeline",
			"output_key": "pipeline"
		},
	}
*/
type Pipeline struct {
	Options   PipelineOptions  `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
PipelineOptions contains custom options for the Pipeline processor:

	Processors:
		array of processors applied to the data
*/
type PipelineOptions struct {
	Processors []config.Config
}

// ApplyBatch processes a slice of encapsulated data with the Pipeline processor. Conditions are optionally applied to the data to enable processing.
func (p Pipeline) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process pipeline: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process pipeline: %v", err)
	}

	return capsules, nil
}

/*
Apply processes encapsulated data with the Pipeline processor.

Applicators only accept encapsulated data, so when processing
JSON the input value is converted from Result to its
string representation to bytes and put into a new capsule.
The conversion from Result to string is safe for strings and
objects, but not arrays (e.g., ["foo","bar"]).

If the input is an array, then an error is raised; the
input should be run through the ForEach processor (which
can encapsulate the Pipeline processor).
*/
func (p Pipeline) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	applicators, err := MakeApplicators(p.Options.Processors)
	if err != nil {
		return capsule, fmt.Errorf("process pipeline: processors %+v: %v", p.Options.Processors, err)
	}

	if p.InputKey != "" && p.OutputKey != "" {
		result := capsule.Get(p.InputKey)
		if result.IsArray() {
			return capsule, fmt.Errorf("process pipeline: inputkey %s: %v", p.InputKey, errPipelineArrayInput)
		}

		newCap := config.NewCapsule()
		newCap.SetData([]byte(result.String()))

		newCap, err = Apply(ctx, newCap, applicators...)
		if err != nil {
			return capsule, fmt.Errorf("process pipeline: %v", err)
		}

		if err := capsule.Set(p.OutputKey, newCap.Data()); err != nil {
			return capsule, fmt.Errorf("process pipeline: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		tmp, err := Apply(ctx, capsule, applicators...)
		if err != nil {
			return capsule, fmt.Errorf("process pipeline: %v", err)
		}

		return tmp, nil
	}

	return capsule, fmt.Errorf("process pipeline: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
}
