package condition

import (
	iconfig "github.com/brexhq/substation/internal/config"
)

type stringConfig struct {
	Object iconfig.Object `json:"object"`

	String string `json:"string"`
}

func (c *stringConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}
