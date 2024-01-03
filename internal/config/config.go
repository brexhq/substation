// package config provides configuration types and functions for Substation.
//
// Any non-backwards compatible changes to the configuration types should be
// accompanied by a version bump.
package config

import (
	"encoding/json"

	"github.com/brexhq/substation/config"
)

type Object struct {
	SourceKey string `json:"source_key"`
	TargetKey string `json:"target_key"`
	BatchKey  string `json:"batch_key"`
}

type AWS struct {
	Region  string `json:"region"`
	RoleARN string `json:"role_arn"`
}

type Metric struct {
	Name        string            `json:"name"`
	Attributes  map[string]string `json:"attributes"`
	Destination config.Config     `json:"destination"`
}

type Request struct {
	Timeout string `json:"Timeout"`
}

type Retry struct {
	Count int `json:"count"`
}

type Batch struct {
	Count    int    `json:"count"`
	Size     int    `json:"size"`
	Duration string `json:"duration"`
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
