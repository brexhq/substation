package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// PipelineArrayInput is returned when the Pipeline processor is configured to process JSON, but the input is an array. Array values are not supported by this processor, instead the input should be run through the ForEach processor (which can encapsulate the Pipeline processor).
const PipelineArrayInput = errors.Error("PipelineArrayInput")

/*
PipelineOptions contains custom options for the Pipeline processor:
	Processors:
		array of processors to apply to the data
*/
type PipelineOptions struct {
	Processors []config.Config
}

/*
Pipeline processes data by applying a series of processors. This processor should be used when data requires complex processing outside of the boundaries of any data structures (see tests for examples). The processor supports these patterns:
	json:
		{"pipeline":"H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA"} >>> {"pipeline":"foo"}
	data:
		H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA >> foo

The processor uses this Jsonnet configuration:
	{
		type: 'pipeline',
		settings: {
			input_key: 'pipeline',
			output_key: 'pipeline',
			options: {
				processors: [
					{
						type: 'base64',
						settings: {
							options: {
								direction: 'from',
							}
						}
					},
					{
						type: 'gzip',
						settings: {
							options: {
								direction: 'from',
							}
						}
					},
				]
			}
		},
	}
*/
type Pipeline struct {
	Condition condition.OperatorConfig `json:"condition"`
	Options   PipelineOptions          `json:"options"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Pipeline processor. Conditions are optionally applied on the bytes to enable processing.
func (p Pipeline) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %v: %v", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %v: %v", p, err)
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("slicer: %v", err)
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

/*
Byte processes bytes with the Pipeline processor.

Process Byters only accept bytes, so when processing
JSON the input value is converted from Result to its
string representation to bytes. The conversion from
Result to string is safe for strings and objects, but
not arrays (e.g., ["foo","bar"]).

If the input is an array, then an error is raised; the
input should be run through the ForEach processor (which
can encapsulate the Pipeline processor).
*/
func (p Pipeline) Byte(ctx context.Context, data []byte) ([]byte, error) {
	byters, err := MakeAllByters(p.Options.Processors)
	if err != nil {
		return nil, fmt.Errorf("byter settings %v: %v", p, err)
	}

	if p.InputKey != "" && p.OutputKey != "" {
		value := json.Get(data, p.InputKey)
		if value.IsArray() {
			return nil, fmt.Errorf("byter settings %v: %v", p, PipelineArrayInput)
		}

		tmp, err := Byte(ctx, byters, []byte(value.String()))
		if err != nil {
			return nil, fmt.Errorf("byter settings %v: %v", p, err)
		}

		return json.Set(data, p.OutputKey, tmp)
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		tmp, err := Byte(ctx, byters, data)
		if err != nil {
			return nil, fmt.Errorf("byter settings %v: %v", p, err)
		}

		return tmp, nil
	}

	return nil, fmt.Errorf("byter settings %v: %v", p, ProcessorInvalidSettings)
}
