package condition

import (
	"github.com/brexhq/substation/internal/json"
)

// JSONSchema implements the Inspector interface for evaluating JSON schemas. More information is available in the README.
type JSONSchema struct {
	Schema []struct {
		Key  string `mapstructure:"key"`
		Type string `mapstructure:"type"`
	} `mapstructure:"schema"`
	Negate bool `mapstructure:"negate"`
}

// Inspect evaluates the JSON object against a provided schema.
func (c JSONSchema) Inspect(data []byte) (output bool, err error) {
	matched := true

	for _, schema := range c.Schema {
		v := json.Get(data, schema.Key)
		vtype := json.Types[v.Type]

		// Null values don't exist in the JSON
		// 	and cannot be validated
		if vtype == "Null" {
			continue
		}

		// validates that values are one of ...
		// 	string OR string array
		// 	number OR number array
		// 	boolean OR boolean array
		// 	pre-formatted JSON
		if v.IsArray() && vtype+"/array" != schema.Type {
			matched = false
		} else if vtype != schema.Type {
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

// JSONValid implements the Inspector interface for evaluating the validity of JSON data. More information is available in the README.
type JSONValid struct {
	Negate bool `mapstructure:"negate"`
}

// Inspect evaluates data as a valid JSON object.
func (c JSONValid) Inspect(data []byte) (output bool, err error) {
	matched := json.Valid(data)

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
