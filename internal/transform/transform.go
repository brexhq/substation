package transform

import (
	"context"
	"fmt"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

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
		return nil, fmt.Errorf("transform settings %v: %v", cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}
