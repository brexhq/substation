package condition

import (
	"fmt"

	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type metaConfig struct {
	Conditions []config.Config `json:"conditions"`

	Object iconfig.Object `json:"object"`
}

func (c *metaConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaConfig) Validate() error {
	if len(c.Conditions) == 0 {
		return fmt.Errorf("conditions: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}
