package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errpipelineArrayInput is returned when the pipeline processor is configured to process JSON, but the input is an array. Array values are not supported by this processor, instead the input should be run through the ForEach processor (which can encapsulate the pipeline processor).
const errpipelineArrayInput = errors.Error("input is an array")

type pipeline struct {
	process
	Options pipelineOptions `json:"options"`
}

type pipelineOptions struct {
	Processors []config.Config
}

// Close closes resources opened by the pipeline processor.
func (p pipeline) Close(context.Context) error {
	return nil
}

func (p pipeline) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process pipeline: %v", err)
	}

	return capsules, nil
}

/*
Apply processes encapsulated data with the pipeline processor.

Applicators only accept encapsulated data, so when processing
JSON the input value is converted from Result to its
string representation to bytes and put into a new capsule.
The conversion from Result to string is safe for strings and
objects, but not arrays (e.g., ["foo","bar"]).

If the input is an array, then an error is raised; the
input should be run through the ForEach processor (which
can encapsulate the pipeline processor).
*/
func (p pipeline) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	applicators, err := MakeApplicators(p.Options.Processors...)
	if err != nil {
		return capsule, fmt.Errorf("process pipeline: processors %+v: %v", p.Options.Processors, err)
	}

	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key)
		if result.IsArray() {
			return capsule, fmt.Errorf("process pipeline: inputkey %s: %v", p.Key, errpipelineArrayInput)
		}

		newCapsule := config.NewCapsule()
		newCapsule.SetData([]byte(result.String()))

		newCapsule, err = Apply(ctx, newCapsule, applicators...)
		if err != nil {
			return capsule, fmt.Errorf("process pipeline: %v", err)
		}

		if err := capsule.Set(p.SetKey, newCapsule.Data()); err != nil {
			return capsule, fmt.Errorf("process pipeline: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
		tmp, err := Apply(ctx, capsule, applicators...)
		if err != nil {
			return capsule, fmt.Errorf("process pipeline: %v", err)
		}

		return tmp, nil
	}

	return capsule, fmt.Errorf("process pipeline: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}
