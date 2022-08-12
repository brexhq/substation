package condition

import (
	"github.com/brexhq/substation/config"
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
		Key  string `json:"key"`
		Type string `json:"type"`
	} `json:"schema"`
	Negate bool `json:"negate"`
}

// Inspect evaluates encapsulated data with the JSONSchema inspector.
func (c JSONSchema) Inspect(cap config.Capsule) (output bool, err error) {
	matched := true

	for _, schema := range c.Schema {
		result := cap.Get(schema.Key)
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
	Negate bool `json:"negate"`
}

// Inspect evaluates encapsulated data with the JSONValid inspector.
func (c JSONValid) Inspect(cap config.Capsule) (output bool, err error) {
	matched := json.Valid(cap.GetData())

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
