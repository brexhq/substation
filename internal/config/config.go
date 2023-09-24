// package config provides configuration types and functions for Substation.
//
// Any non-backwards compatible changes to the configuration types should be
// accompanied by a version bump.
package config

import (
	"encoding/json"
)

type Object struct {
	Key    string `json:"key"`
	SetKey string `json:"set_key"`
}

type AWS struct {
	Region        string `json:"region"`
	AssumeRoleARN string `json:"assume_role_arn"`
}

type Request struct {
	Timeout string `json:"Timeout"`
}

type Retry struct {
	Count int `json:"count"`
}

// Buffer should be used by any transform that supports buffering data
// with internal/aggregate.
type Buffer struct {
	Count    int    `json:"count"`
	Size     int    `json:"size"`
	Duration string `json:"duration"`
	Key      string `json:"key"`
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
