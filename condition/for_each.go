package condition

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errForEachInvalidType is returned when the forEach inspector is configured with an invalid type.
const errForEachInvalidType = errors.Error("invalid type")

// forEach evaluates conditions by iterating and applying an inspector to each element in a JSON array.
//
// This inspector supports the object handling pattern.
type forEach struct {
	condition
	Options forEachOptions `json:"options"`
}

type forEachOptions struct {
	// Type determines the method of combining results from the inspector.
	//
	// Must be one of:
	//	- none: none of the elements match the condition
	//	- any: at least one of the elements match the condition
	//	- all: all of the elements match the condition
	Type string `json:"type"`
	// Inspector is the condition applied to each element.
	Inspector config.Config `json:"inspector"`
}

// Inspect evaluates encapsulated data with the Content inspector.
func (c forEach) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	conf, err := gojson.Marshal(c.Options.Inspector)
	if err != nil {
		return false, fmt.Errorf("condition: for_each: %w", err)
	}

	var condition config.Config
	if err = gojson.Unmarshal(conf, &condition); err != nil {
		return false, fmt.Errorf("condition: for_each: %w", err)
	}

	inspector, err := InspectorFactory(condition)
	if err != nil {
		return false, fmt.Errorf("condition: for_each: %w", err)
	}

	var results []bool
	for _, res := range capsule.Get(c.Key).Array() {
		tmpCapule := config.NewCapsule()
		tmpCapule.SetData([]byte(res.String()))

		inspected, err := inspector.Inspect(ctx, tmpCapule)
		if err != nil {
			return false, fmt.Errorf("condition: for_each: %w", err)
		}
		results = append(results, inspected)
	}

	total := len(results)
	matched := 0
	for _, v := range results {
		if v {
			matched++
		}
	}

	switch c.Options.Type {
	case "any":
		output = matched > 0
	case "all":
		output = total == matched
	case "none":
		output = matched == 0
	default:
		return false, fmt.Errorf("condition for_each: type %q: %v", c.Options.Type, errForEachInvalidType)
	}

	if c.Negate {
		return !output, nil
	}

	return output, nil
}
