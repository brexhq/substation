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
		return newTformBatch(cfg)
	case "transfer":
		return newTformTransfer(cfg)
	default:
		return nil, fmt.Errorf("transform settings %v: %w", cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}
