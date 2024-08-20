package condition

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/internal/errors"
	"github.com/brexhq/substation/v2/message"
)

type Inspector interface {
	Inspect(context.Context, *message.Message) (bool, error)
}

func New(ctx context.Context, cfg config.Config) (Inspector, error) {
	switch cfg.Type {
	case "all", "meta_all":
		return newMetaAll(ctx, cfg)
	case "any", "meta_any":
		return newMetaAny(ctx, cfg)
	case "none", "meta_none":
		return newMetaNone(ctx, cfg)
	case "string_contains":
		return newStringContains(ctx, cfg)
	case "string_match":
		return newStringMatch(ctx, cfg)
	default:
		return nil, fmt.Errorf("condition %s: %w", cfg.Type, errors.ErrInvalidFactoryInput)
	}
}
