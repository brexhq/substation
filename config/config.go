// Package config provides structures for building configurations.
package config

import (
	"encoding/json"
)

// Config is a template used by Substation interface factories to produce new
// instances. Type refers to the type of instance and Settings contains options
// used in the instance. Examples of this are found in the condition and transforms
// packages.
type Config struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings"`
}

func (c Config) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}
