package process

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/itchyny/gojq"
)

// errJq is returned when the jq query generates no output.
const errJqNoOutputGenerated = errors.Error("no output generated")

// jq processes data by applying jq queries to it.
//
// This processor supports the object handling pattern.
type _jq struct {
	process
	Options _jqOptions `json:"options"`
}

type _jqOptions struct {
	Query string `json:"query"`
}

// String returns the processor settings as an jq.
func (p _jq) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p _jq) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _jq) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
func (p _jq) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	query, err := gojq.Parse(p.Options.Query)
	if err != nil {
		return capsule, err
	}

	var i interface{}
	if err := gojson.Unmarshal(capsule.Data(), &i); err != nil {
		return capsule, fmt.Errorf("process: jq: %v", err)
	}

	var arr []interface{}
	iter := query.Run(i)

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return capsule, fmt.Errorf("process: jq: %v", err)
		}

		arr = append(arr, v)
	}

	var b []byte
	switch len(arr) {
	case 0:
		err = errJqNoOutputGenerated
	case 1:
		b, err = gojson.Marshal(arr[0])
		capsule.SetData(b)
	default:
		b, err = gojson.Marshal(arr)
		capsule.SetData(b)
	}

	if err != nil {
		return capsule, fmt.Errorf("process: jq: %v", err)
	}

	return capsule, nil
}
