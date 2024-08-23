package condition

import (
	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type metaConfig struct {
	Inspectors []config.Config `json:"inspectors"`

	Object iconfig.Object `json:"object"`
}

func (c *metaConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}
