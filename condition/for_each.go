package condition

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errForEachInvalidType is returned when the ForEach inspector is configured with an invalid type.
const errForEachInvalidType = errors.Error("invalid type")

/*
ForEach evaluates conditions by iterating and applying a condition to each element in a JSON array.

The inspector has these settings:

	Options:
		Condition inspector to be applied to all array elements.
	Type:
		Method of combining the results of the conditions evaluated.
		Must be one of:
			none: none of the elements must match the condition
			any: at least one of the elements must match the condition
			all: all of the elements must match the condition
	Key:
		JSON key-value to retrieve for inspection
	Negate (optional):
		If set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

When loaded with a factory, the inspector uses this JSON configuration:

	{
		"options": {
			"type": "strings",
			"settings": {
				"function": "endswith",
				"expression": "@example.com"
			}
		},
		"type": "all",
		"key:": "input",
		"negate": false
	}
*/
type ForEach struct {
	Options ForEachOptions `json:"options"`
	Type    string         `json:"type"`
	Key     string         `json:"key"`
	Negate  bool           `json:"negate"`
}

/*
ForEachOptions contains custom options for the ForEach processor:

	Inspector:
		condition applied to the data
*/
type ForEachOptions struct {
	Inspector config.Config `json:"inspector"`
}

// Inspect evaluates encapsulated data with the Content inspector.
func (c ForEach) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
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

	switch c.Type {
	case "any":
		output = matched > 0
	case "all":
		output = total == matched
	case "none":
		output = matched == 0
	default:
		return false, fmt.Errorf("condition for_each: type %q: %v", c.Type, errForEachInvalidType)
	}

	if c.Negate {
		return !output, nil
	}

	return output, nil
}
