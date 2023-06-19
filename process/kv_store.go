//go:build !wasm

package process

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/kv"
)

// kvStore processes data by retrieving values from and putting values into
// key-value (KV) stores.
//
// This processor supports the object handling pattern.
type procKVStore struct {
	process
	Options procKVStoreOptions `json:"options"`

	kvStore kv.Storer
}

type procKVStoreOptions struct {
	// Type determines the action applied to the KV store.
	//
	// Must be one of:
	//
	// - get: value is retrieved from the store
	//
	// - set: value is put into the store
	Type string `json:"type"`
	// Prefix is prepended to either the Key (in the case of get)
	// or the SetKey (in the case of set) and is intended to simplify
	// data management within a KV store.
	//
	// This is optional and defaults to an empty string.
	Prefix string `json:"prefix"`
	// TTLKey retrieves a value from an object that is used as the time-to-live (TTL)
	// of the item set into the KV store. This value must be an integer that represents
	// the epoch time (in seconds) when the item will be evicted from the store.
	//
	// This is optional and defaults to using no TTL when setting items into the store.
	TTLKey string `json:"ttl_key"`
	// OffsetTTL is an offset (in seconds) used to determine the time-to-live (TTL)
	// of the item set into the KV store. If TTLKey is configured, then this value is
	// added to the TTL value retrieved from the object. If TTLKey is not used, then this
	// value is added to the current time.
	//
	// For example, if TTLKey is not set and the offset is 86400 (1 day), then the value
	// will be evicted from the store if more than 1 day has elapsed.
	//
	// This is optional and defaults to using no TTL when setting values into the store.
	OffsetTTL int64 `json:"offset_ttl"` // TODO(v1.0.0): rename to TTLOffset?
	// KVOptions determine the type of KV store used by the processor. Refer to internal/kv
	// for more information.
	KVOptions config.Config `json:"kv_options"`
}

// Create a new pipeline processor.
func newProcKVStore(ctx context.Context, cfg config.Config) (p procKVStore, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procKVStore{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procKVStore{}, err
	}

	//  validate option.type
	if !slices.Contains(
		[]string{
			"get",
			"set",
		},
		p.Options.Type) {
		return procKVStore{}, fmt.Errorf("process: kv_store: type %q: %v", p.Options.Type, errors.ErrInvalidOption)
	}

	// only supports objects, fail if there are no keys
	if p.Key == "" || p.SetKey == "" {
		return procKVStore{}, fmt.Errorf("process: kv_store: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	p.kvStore, err = kv.Get(p.Options.KVOptions)
	if err != nil {
		return procKVStore{}, fmt.Errorf("process: kv_store: %v", err)
	}

	// lazy load the KV store
	if !p.kvStore.IsEnabled() {
		if err := p.kvStore.Setup(ctx); err != nil {
			return procKVStore{}, fmt.Errorf("process: kv_store: %v", err)
		}
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procKVStore) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procKVStore) Close(ctx context.Context) error {
	if p.IgnoreClose {
		return nil
	}

	if p.kvStore.IsEnabled() {
		if err := p.kvStore.Close(); err != nil {
			return fmt.Errorf("close: kv_store: %v", err)
		}
	}

	return nil
}

// Stream processes a pipeline of capsules with the processor.
func (p procKVStore) Stream(ctx context.Context, in, out *config.Channel) error {
	return streamApply(ctx, in, out, p)
}

// Batch processes one or more capsules with the processor.
func (p procKVStore) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p)
}

// Apply processes a capsule with the processor.
func (p procKVStore) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	if ok, err := p.operator.Operate(ctx, capsule); err != nil {
		return capsule, fmt.Errorf("process: kv_store: %v", err)
	} else if !ok {
		return capsule, nil
	}

	switch p.Options.Type {
	case "get":
		key := capsule.Get(p.Key).String()
		if p.Options.Prefix != "" {
			key = fmt.Sprint(p.Options.Prefix, ":", key)
		}

		val, err := p.kvStore.Get(ctx, key)
		if err != nil {
			return capsule, fmt.Errorf("process: kv_store: %v", err)
		}

		if err := capsule.Set(p.SetKey, val); err != nil {
			return capsule, fmt.Errorf("process: kv_store: %v", err)
		}

		return capsule, nil
	case "set":
		key := capsule.Get(p.SetKey).String()
		if p.Options.Prefix != "" {
			key = fmt.Sprint(p.Options.Prefix, ":", key)
		}

		//nolint: nestif // ignore nesting complexity
		if p.Options.TTLKey != "" && p.Options.OffsetTTL != 0 {
			ttl := capsule.Get(p.Options.TTLKey).Int() + p.Options.OffsetTTL
			if err := p.kvStore.SetWithTTL(ctx, key, capsule.Get(p.Key).String(), ttl); err != nil {
				return capsule, fmt.Errorf("process: kv_store: %v", err)
			}
		} else if p.Options.OffsetTTL != 0 {
			ttl := time.Now().Add(time.Duration(p.Options.OffsetTTL) * time.Second).Unix()
			if err := p.kvStore.SetWithTTL(ctx, key, capsule.Get(p.Key).String(), ttl); err != nil {
				return capsule, fmt.Errorf("process: kv_store: %v", err)
			}
		} else {
			if err := p.kvStore.Set(ctx, key, capsule.Get(p.Key).String()); err != nil {
				return capsule, fmt.Errorf("process: kv_store: %v", err)
			}
		}

		return capsule, nil
	default:
		return capsule, fmt.Errorf("process: kv_store: %v", errors.ErrInvalidOption)
	}
}
