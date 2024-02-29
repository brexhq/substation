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
	// SourceKey retrieves a value from a JSON object.
	SourceKey string `json:"source_key"`
	// TargetKey place a value into a JSON object.
	TargetKey string `json:"target_key"`
	// BatchKey retrieves a value from a JSON object that is used to organize
	// batched data (internal/aggregate).
	BatchKey string `json:"batch_key"`
}

type AWS struct {
	// Region is the AWS region that the AWS client will use.
	Region string `json:"region"`
	// RoleARN is the AWS IAM role that the AWS client will assume.
	RoleARN string `json:"role_arn"`
}

type Metric struct {
	// Name is the name of the metric.
	Name string `json:"name"`
	// Attributes are key-value pairs that are associated with the metric.
	Attributes map[string]string `json:"attributes"`
	// Destination is the metrics destination that the metric will be sent to (internal/metrics).
	Destination config.Config `json:"destination"`
}

type Request struct {
	// Timeout is the amount of time that the request will wait before timing out.
	Timeout string `json:"Timeout"`
}

type Retry struct {
	// Count is the maximum number of times that the action will be retried.
	Count int `json:"count"`
	// ErrorMessages are regular expressions that match error messages and determine
	// if the action should be retried.
	ErrorMessages []string `json:"error_messages"`
}

type Batch struct {
	// Count is the maximum number of records that can be batched.
	Count int `json:"count"`
	// Size is the maximum size of the batch in bytes.
	Size int `json:"size"`
	// Duration is the maximum amount of time that records can be batched for.
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
