package process

import (
	"context"
	"fmt"
	"time"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/cache/backend"
	"github.com/brexhq/substation/internal/errors"
)

const errCacheInvalidType = errors.Error("invalid type")

type _cache struct {
	process
	Options _cacheSetOptions `json:"options"`
}

type _cacheSetOptions struct {
	Type   string        `json:"type"`
	Prefix string        `json:"prefix"`
	SetTTL int           `json:"set_ttl"`
	Cache  config.Config `json:"cache"`
}

// String returns the processor settings as an object.
func (p _cache) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p _cache) Close(ctx context.Context) error {
	if p.IgnoreClose {
		return nil
	}

	cache, err := backend.Factory(p.Options.Cache)
	if err != nil {
		return fmt.Errorf("close cache: %v", err)
	}

	if cache.IsEnabled() {
		if err := cache.Close(); err != nil {
			return fmt.Errorf("close cache: %v", err)
		}
	}

	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _cache) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
func (p _cache) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports objects, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process: cache: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	cache, err := backend.Factory(p.Options.Cache)
	if err != nil {
		return capsule, fmt.Errorf("process: cache: %v", err)
	}

	// lazy load the cache
	if !cache.IsEnabled() {
		if err := cache.Setup(ctx); err != nil {
			return capsule, fmt.Errorf("process: cache: %v", err)
		}
	}

	switch p.Options.Type {
	case "get":
		key := capsule.Get(p.Key).String()
		if p.Options.Prefix != "" {
			key = fmt.Sprint(p.Options.Prefix, ":", key)
		}

		result, err := cache.Get(ctx, key)
		if err != nil {
			return capsule, fmt.Errorf("process: cache: %v", err)
		}

		if err := capsule.Set(p.SetKey, result); err != nil {
			return capsule, fmt.Errorf("process: cache: %v", err)
		}

		return capsule, nil
	case "set":
		key := capsule.Get(p.SetKey).String()
		if p.Options.Prefix != "" {
			key = fmt.Sprint(p.Options.Prefix, ":", key)
		}

		if p.Options.SetTTL == 0 {
			if err := cache.Put(ctx, key, capsule.Get(p.Key).String()); err != nil {
				return capsule, fmt.Errorf("process: cache: %v", err)
			}
		} else {
			ttl := time.Now().Add(time.Duration(p.Options.SetTTL) * time.Second).Unix()
			if err := cache.PutWithTTL(ctx, key, capsule.Get(p.Key).String(), ttl); err != nil {
				return capsule, fmt.Errorf("process: cache: %v", err)
			}
		}

		return capsule, nil
	default:
		return capsule, fmt.Errorf("process: cache: %v", errCacheInvalidType)
	}
}
