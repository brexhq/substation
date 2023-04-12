package process

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/itchyny/gojq"
)

// errJqNoOutputGenerated is returned when the jq query generates no output.
const errJqNoOutputGenerated = errors.Error("no output generated")

// jq processes data by applying jq queries.
//
// This processor supports the data handling pattern.
type procJQ struct {
	process
	Options procJQOptions `json:"options"`

	query *gojq.Query
}

type procJQOptions struct {
	// Query is the jq query applied to data.
	Query string `json:"query"`
}

// Create a new join processor.
func newProcJQ(cfg config.Config) (p procJQ, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procJQ{}, err
	}

	p.operator, err = condition.NewOperator(p.Condition)
	if err != nil {
		return procJQ{}, err
	}

	if p.Options.Query == "" {
		return procJQ{}, fmt.Errorf("process: jq: query %q: %v", p.Options.Query, errors.ErrMissingRequiredOption)
	}

	p.query, err = gojq.Parse(p.Options.Query)
	if err != nil {
		return procJQ{}, err
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procJQ) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procJQ) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procJQ) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes encapsulated data with the processor.
func (p procJQ) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	var i interface{}
	if err := gojson.Unmarshal(capsule.Data(), &i); err != nil {
		return capsule, fmt.Errorf("process: jq: %v", err)
	}

	var arr []interface{}
	iter := p.query.RunWithContext(ctx, i)

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

	var err error
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
