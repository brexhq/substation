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
	// OffsetTTL is an offset (in seconds) used to determine the time-to-live (TTL)
	// of the value set into the KV store. TTL is calculated based on the current
	// time plus the offset.
	//
	// For example, if the offset is 86400 (1 day), then the value will either be
	// evicted from the store or ignored on retrieval if more than 1 day has elapsed
	// since it was placed into the store.
	//
	// This is optional and defaults to using no TTL when setting values into the store.
	OffsetTTL int `json:"offset_ttl"`
	// KVOptions determine the type of KV store used by the processor. Refer to internal/kv
	// for more information.
	KVOptions config.Config `json:"kv_options"`
}

// Create a new pipeline processor.
func newProcKVStore(cfg config.Config) (p procKVStore, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procKVStore{}, err
	}

	p.operator, err = condition.NewOperator(p.Condition)
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

	ctx := context.Background()
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

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procKVStore) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procKVStore) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
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

		if p.Options.OffsetTTL == 0 {
			if err := p.kvStore.Set(ctx, key, capsule.Get(p.Key).String()); err != nil {
				return capsule, fmt.Errorf("process: kv_store: %v", err)
			}
		} else {
			ttl := time.Now().Add(time.Duration(p.Options.OffsetTTL) * time.Second).Unix()
			if err := p.kvStore.SetWithTTL(ctx, key, capsule.Get(p.Key).String(), ttl); err != nil {
				return capsule, fmt.Errorf("process: kv_store: %v", err)
			}
		}

		return capsule, nil
	default:
		return capsule, fmt.Errorf("process: kv_store: %v", errors.ErrInvalidOption)
	}
}
