package condition

import (
	"github.com/brexhq/substation/internal/json"
)

/*
JSONSchema evaluates JSON objects against a schema.

The inspector has these settings:
	Schema.Key:
		the JSON key-value to retrieve for inspection
	Schema.Type:
		the value type used during inspection of the Schema.Key
		must be one of:
			string
			number (float, int)
			boolean (true, false)
			json
	Negate (optional):
		if set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

The inspector supports these patterns:
	json:
		{"foo":"foo","bar":123} == string,number

The inspector uses this Jsonnet configuration:
	{
		type: 'json_schema',
		settings: {
			schema: [
				{
					key: "foo",
					type: "string",
				},
				{
					key: "bar",
					type: "number",
				}
			],
		},
	}
*/
type JSONSchema struct {
	Schema []struct {
		Key  string `mapstructure:"key"`
		Type string `mapstructure:"type"`
	} `mapstructure:"schema"`
	Negate bool `mapstructure:"negate"`
}

// Inspect evaluates data with the JSONSchema inspector.
func (c JSONSchema) Inspect(data []byte) (output bool, err error) {
	matched := true

	for _, schema := range c.Schema {
		value := json.Get(data, schema.Key)
		vtype := json.Types[value.Type]

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
		if value.IsArray() && vtype+"/array" != schema.Type {
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

/*
JSONValid evaluates JSON objects for validity.

The inspector has these settings:
	Negate (optional):
		if set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

The inspector supports these patterns:
	json:
		{"foo":"foo","bar":123} == valid
		foo == invalid

The inspector uses this Jsonnet configuration:
	{
		type: 'json_valid',
	}
*/
type JSONValid struct {
	Negate bool `mapstructure:"negate"`
}

// Inspect evaluates data with the JSONValid inspector.
func (c JSONValid) Inspect(data []byte) (output bool, err error) {
	matched := json.Valid(data)

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
