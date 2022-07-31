package config

import "encoding/json"

// Config is a template used by Substation interface factories to produce new instances from JSON configurations. Type refers to the type of instance and Settings contains options used in the instance. Examples of this are found in process.ByterFactory and condition.InspectorFactory.
type Config struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings"`
}

// Decode marshals and unmarshals the input structure into the output structure using the standard library's json package. This should be used when decoding JSON configurations (i.e., Config) in Substation interface factories (e.g., process.ByterFactory, condition.InspectorFactory).
func Decode(input interface{}, output interface{}) error {
	b, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, output)
}
