package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

// TransformInvalidFactoryConfig is returned when an unsupported Transform is referenced in Factory.
const TransformInvalidFactoryConfig = errors.Error("TransformInvalidFactoryConfig")

// Transform is an interface for transforming data as it moves from a source to a sink. Transforms read bytes from and write bytes to channels, may optionally modify bytes, and are interruptable via an anonymous struct channel.
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
