package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

// TransformInvalidFactoryConfig is used when an unsupported Transform is referenced in Factory.
const TransformInvalidFactoryConfig = errors.Error("TransformInvalidFactoryConfig")

// Transform is the interface used by all Substation transforms. Transforms read channels of bytes, can optionally write channels of bytes, and are interruptable via an anonymous struct channel.
type Transform interface {
	Transform(context.Context, <-chan []byte, chan<- []byte, chan struct{}) error
}

// Factory loads Transforms from a Config. This is the recommended function for retrieving ready-to-use Transforms.
func Factory(cfg config.Config) (Transform, error) {
	switch t := cfg.Type; t {
	case "process":
		var t Process
		config.Decode(cfg.Settings, &t)
		return &t, nil
	case "transfer":
		var t Transfer
		config.Decode(cfg.Settings, &t)
		return &t, nil
	default:
		return nil, fmt.Errorf("err retrieving %s from factory: %v", t, TransformInvalidFactoryConfig)
	}
}
