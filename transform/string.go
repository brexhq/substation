package transform

import (
	"fmt"

	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

type strCaseConfig struct {
	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *strCaseConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *strCaseConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func strCaptureGetBytesMatch(match [][]byte) []byte {
	if len(match) > 1 {
		return match[len(match)-1]
	}

	return nil
}

func strCaptureGetStringMatch(match []string) string {
	if len(match) > 1 {
		return match[len(match)-1]
	}

	return ""
}
