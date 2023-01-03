package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// join processes data by joinenating values in an object array.
//
// This processor supports the object handling pattern.
type procJoin struct {
	process
	Options procJoinOptions `json:"options"`
}

type procJoinOptions struct {
	// Separator is the string that joins data from the array.
	Separator string `json:"separator"`
}

// String returns the processor settings as an object.
func (p procJoin) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procJoin) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procJoin) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes encapsulated data with the processor.
func (p procJoin) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Separator == "" {
		return capsule, fmt.Errorf("process: join: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// only supports JSON, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process: join: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	// data is processed by retrieving and iterating the
	// array (Key) containing string values and joining
	// each one with the separator string
	//
	// root:
	// 	{"join":["foo","bar","baz"]}
	// joinenated:
	// 	{"join:"foo.bar.baz"}
	var value string
	result := capsule.Get(p.Key)
	for i, res := range result.Array() {
		value += res.String()
		if i != len(result.Array())-1 {
			value += p.Options.Separator
		}
	}

	if err := capsule.Set(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process: join: %v", err)
	}

	return capsule, nil
}
