package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errInvalidFactoryInput is returned when an unsupported Transform is referenced in Factory.
const errInvalidFactoryInput = errors.Error("invalid factory input")

// Transformer is an interface for transforming data as it moves from a source to a sink. Transformers read capsules from and write capsules to channels, may optionally modify bytes, and are interruptable.
type Transformer interface {
	Transform(context.Context, *config.Channel, *config.Channel) error
}

// Factory returns a configured Transformer from a config. This is the recommended method for retrieving ready-to-use Transformers.
func Factory(cfg config.Config) (Transformer, error) {
	switch t := cfg.Type; t {
	case "batch":
		var t Batch
		_ = config.Decode(cfg.Settings, &t)
		return &t, nil
	case "transfer":
		var t Transfer
		_ = config.Decode(cfg.Settings, &t)
		return &t, nil
	default:
		return nil, fmt.Errorf("transform settings %v: %v", cfg.Settings, errInvalidFactoryInput)
	}
}
