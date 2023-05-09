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
func New(ctx context.Context, cfg config.Config) (Transformer, error) {
	switch t := cfg.Type; t {
	case "batch":
		return newTformBatch(ctx, cfg)
	case "stream":
		return newTformStream(ctx, cfg)
	case "noop":
		fallthrough
	// TODO(v1.0.0): remove and replace with noop(?)
	case "transfer":
		return newTformTransfer(ctx, cfg)
	default:
		return nil, fmt.Errorf("transform settings %v: %v", cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}
