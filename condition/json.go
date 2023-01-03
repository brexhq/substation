package condition

import (
	"context"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
)

// jsonSchema evaluates objects against a minimal schema parser.
//
// This inspector supports the object handling pattern.
type inspJSONSchema struct {
	condition
	Options inspJSONSchemaOptions `json:"options"`
}

type inspJSONSchemaOptions struct {
	Schema []struct {
		// Key is the JSON key to retrieve for inspection.
		Key string `json:"key"`
		// Type is the expected value type for Key.
		//
		// Must be one of:
		//	- string
		//	- number (float, int)
		//	- boolean (true, false)
		//	- json
		Type string `json:"type"`
	} `json:"schema"`
}

func (c inspJSONSchema) String() string {
	return toString(c)
}

// Inspect evaluates encapsulated data with the jsonSchema inspector.
func (c inspJSONSchema) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	matched := true

	for _, schema := range c.Options.Schema {
		result := capsule.Get(schema.Key)
		rtype := json.Types[result.Type]

		// Null values don't exist in the JSON
		// 	and cannot be validated
		if rtype == "Null" {
			continue
		}

		// validates that values are one of ...
		// 	string OR string array
		// 	number OR number array
		// 	boolean OR boolean array
		// 	pre-formatted JSON
		if result.IsArray() && rtype+"/array" != schema.Type {
			matched = false
		} else if rtype != schema.Type {
			matched = false
		}

		// break the loop on the first indication that the JSON does not match the schema
		if !matched {
			break
		}
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}

// jsonValid evaluates objects for validity.
//
// This inspector supports the object handling pattern.
type inspJSONValid struct {
	condition
}

func (c inspJSONValid) String() string {
	return toString(c)
}

// Inspect evaluates encapsulated data with the jsonValid inspector.
func (c inspJSONValid) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	matched := json.Valid(capsule.Data())

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
