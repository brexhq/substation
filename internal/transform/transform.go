package transform

import (
	"context"
	"fmt"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errInvalidFactoryInput is returned when an unsupported Transform is referenced in Factory.
const errInvalidFactoryInput = errors.Error("invalid factory input")

type transformer interface {
	Transform(context.Context, *sync.WaitGroup, *config.Channel, *config.Channel) error
}

// Make returns a configured transform from a transform configuration.
func Make(cfg config.Config) (transformer, error) {
	switch t := cfg.Type; t {
	case "batch":
		var t _batch
		_ = config.Decode(cfg.Settings, &t)
		return &t, nil
	case "transfer":
		var t _transfer
		_ = config.Decode(cfg.Settings, &t)
		return &t, nil
	default:
		return nil, fmt.Errorf("transform settings %v: %v", cfg.Settings, errInvalidFactoryInput)
	}
}
