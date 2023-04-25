package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

// errPipelineArrayInput is returned when the pipeline processor is configured to process JSON, but the input is an array. Array values are not supported by this processor, instead the input should be run through the ForEach processor (which can encapsulate the pipeline processor).
var errPipelineArrayInput = fmt.Errorf("input is an array")

// pipeline processes data by applying a series of processors.
//
// This processor supports the data and object handling patterns.
type procPipeline struct {
	process
	Options procPipelineOptions `json:"options"`

	appliers []Applier
}

type procPipelineOptions struct {
	// Processors applied in series to the data.
	Processors []config.Config `json:"processors"`
}

// Create a new pipeline processor.
func newProcPipeline(ctx context.Context, cfg config.Config) (p procPipeline, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procPipeline{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procPipeline{}, err
	}

	p.appliers, err = NewAppliers(ctx, p.Options.Processors...)
	if err != nil {
		return procPipeline{}, fmt.Errorf("process: pipeline: processors %+v: %v", p.Options.Processors, err)
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procPipeline) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procPipeline) Close(context.Context) error {
	return nil
}

// Stream processes a pipeline of capsules with the processor.
func (p procPipeline) Stream(ctx context.Context, in, out *config.Channel) error {
	return streamApply(ctx, in, out, p)
}

// Batch processes one or more capsules with the processor.
func (p procPipeline) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p)
}

// Apply processes a capsule with the processor.
//
// Appliers only accept encapsulated data, so when processing
// objects the data is converted to its string representation to
// bytes and put into a new capsule. The conversion to string is
// safe for strings and objects, but not arrays
// (e.g., ["foo","bar"]).
//
// If the input is an array, then an error is raised; the input
// should be run through the forEach processor (which can
// encapsulate the pipeline processor).
func (p procPipeline) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	if ok, err := p.operator.Operate(ctx, capsule); err != nil {
		return capsule, fmt.Errorf("process: pipeline: %v", err)
	} else if !ok {
		return capsule, nil
	}

	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key)
		if result.IsArray() {
			return capsule, fmt.Errorf("process: pipeline: key %s: %v", p.Key, errPipelineArrayInput)
		}

		newCapsule := config.NewCapsule()
		newCapsule.SetData([]byte(result.String()))

		newCapsule, err := Apply(ctx, newCapsule, p.appliers...)
		if err != nil {
			return capsule, fmt.Errorf("process: pipeline: %v", err)
		}

		if err := capsule.Set(p.SetKey, newCapsule.Data()); err != nil {
			return capsule, fmt.Errorf("process: pipeline: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
		tmp, err := Apply(ctx, capsule, p.appliers...)
		if err != nil {
			return capsule, fmt.Errorf("process: pipeline: %v", err)
		}

		return tmp, nil
	}

	return capsule, fmt.Errorf("process: pipeline: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}
