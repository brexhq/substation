package config

import "encoding/json"

// ConfigAWSAuth is used by functions that require AWS authentication.
type ConfigAWSAuth struct {
	Region     string `json:"region"`
	AssumeRole string `json:"assume_role"`
}

// ConfigRequest is used by functions that make requests over a network.
type ConfigRequest struct {
	Timeout    int `json:"timeout"`
	MaxRetries int `json:"max_retries"`
}

// Decode marshals and unmarshals an input interface into the output interface
// using the standard library's json package. This should be used when decoding
// JSON configurations (i.e., Config) in Substation interface factories.
func Decode(input, output interface{}) error {
	b, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, output)
}