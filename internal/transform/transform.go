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

type Transformer interface {
	Transform(context.Context, *sync.WaitGroup, *config.Channel, *config.Channel) error
}

// New returns a configured Transformer from a transform configuration.
func New(cfg config.Config) (Transformer, error) {
	switch t := cfg.Type; t {
	case "batch":
		var t tformBatch
		_ = config.Decode(cfg.Settings, &t)
		return &t, nil
	case "transfer":
		var t tformTransfer
		_ = config.Decode(cfg.Settings, &t)
		return &t, nil
	default:
		return nil, fmt.Errorf("transform settings %v: %v", cfg.Settings, errInvalidFactoryInput)
	}
}
