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
	SrcKey string `json:"src_key"`
	DstKey string `json:"dst_key"`
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
