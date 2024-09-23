package transform

import (
	"fmt"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type hashConfig struct {
	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *hashConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *hashConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}
