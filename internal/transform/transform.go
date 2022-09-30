package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// TransformInvalidFactoryConfig is returned when an unsupported Transform is referenced in Factory.
const TransformInvalidFactoryConfig = errors.Error("TransformInvalidFactoryConfig")

// Transform is an interface for transforming data as it moves from a source to a sink. Transforms read capsules from and write capsules to channels, may optionally modify bytes, and are interruptable via an anonymous struct channel.
type Transform interface {
	Transform(context.Context, *config.Channel, *config.Channel) error
}

// Factory returns a configured Transform from a config. This is the recommended method for retrieving ready-to-use Transforms.
func Factory(cfg config.Config) (Transform, error) {
	switch t := cfg.Type; t {
	case "batch":
		var t Batch
		config.Decode(cfg.Settings, &t)
		return &t, nil
	// case "transfer":
	// 	var t Transfer
	// 	config.Decode(cfg.Settings, &t)
	// 	return &t, nil
	default:
		return nil, fmt.Errorf("transform settings %v: %v", cfg.Settings, TransformInvalidFactoryConfig)
	}
}
