package condition

import (
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type networkIPConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *networkIPConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}
