package config

import (
	"encoding/json"
)

// ConfigAWSAuth is used by functions that require AWS authentication.
type ConfigAWSAuth struct {
	Region     string `json:"region"`
	AssumeRole string `json:"assume_role"`
}

// ConfigRequest is used by functions that make requests over a network.
type ConfigRequest struct {
	MaxRetries int `json:"max_retries"`
	// Timeout is the maximum amount of time a request can take. This is
	// parsed into a duration using the standard library's time.ParseDuration()
	// function so it supports any valid duration string.
	Timeout string `json:"timeout"`
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
