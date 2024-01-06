package condition

import (
	iconfig "github.com/brexhq/substation/internal/config"
)

type stringConfig struct {
	// Value used for comparison during inspection.
	Value string `json:"value"`

	Object iconfig.Object `json:"object"`
}

func (c *stringConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}
