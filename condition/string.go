package condition

import (
	iconfig "github.com/brexhq/substation/internal/config"
)

type stringConfig struct {
	Object iconfig.Object `json:"object"`

	Value string `json:"value"`
}

func (c *stringConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}
