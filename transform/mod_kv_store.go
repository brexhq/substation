//go:build !wasm

package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"time"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/kv"
	"github.com/brexhq/substation/message"
)

type modKVStoreConfig struct {
	Object configObject `json:"object"`

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
	// TTLOffset is an offset (in seconds) used to determine the time-to-live (TTL)
	// of the item set into the KV store. If TTLKey is configured, then this value is
	// added to the TTL value retrieved from the object. If TTLKey is not used, then this
	// value is added to the current time.
	//
	// For example, if TTLKey is not set and the offset is 86400 (1 day), then the value
	// will be evicted from the store if more than 1 day has elapsed.
	//
	// This is optional and defaults to using no TTL when setting values into the store.
	TTLOffset int64 `json:"ttl_offset"`
	// KVStore determine the type of KV store used by the transform. Refer to internal/kv
	// for more information.
	KVStore config.Config `json:"kv_store"`
	// CloseKVStore determines if the KV store is closed when a control
	// message is received.
	CloseKVStore bool `json:"close_kv_store"`
}

type modKVStore struct {
	conf    modKVStoreConfig
	kvStore kv.Storer
}

func newModKVStore(ctx context.Context, cfg config.Config) (*modKVStore, error) {
	conf := modKVStoreConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_kv_store: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" {
		return nil, fmt.Errorf("transform: new_mod_kv_store: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_kv_store: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Type == "" {
		return nil, fmt.Errorf("transform: new_mod_kv_store: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"get",
			"set",
		},
		conf.Type) {
		return nil, fmt.Errorf("transform: new_mod_kv_store: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	kvStore, err := kv.Get(conf.KVStore)
	if err != nil {
		return nil, fmt.Errorf("transform: new_mod_kv_store: kv_store: %v", err)
	}

	tf := modKVStore{
		conf:    conf,
		kvStore: kvStore,
	}

	if !tf.kvStore.IsEnabled() {
		if err := tf.kvStore.Setup(ctx); err != nil {
			return nil, fmt.Errorf("transform: new_mod_kv_store: kv_store: %v", err)
		}
	}

	return &tf, nil
}

func (tf *modKVStore) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modKVStore) Close(context.Context) error {
	return nil
}

func (tf *modKVStore) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		if !tf.conf.CloseKVStore {
			return []*message.Message{msg}, nil
		}

		if err := tf.kvStore.Close(); err != nil {
			return nil, fmt.Errorf("transform: mod_kv_store: %v", err)
		}

		return []*message.Message{msg}, nil
	}

	if !tf.kvStore.IsEnabled() {
		if err := tf.kvStore.Setup(ctx); err != nil {
			return nil, fmt.Errorf("transform: mod_kv_store: kv_store: %v", err)
		}
	}

	switch tf.conf.Type {
	case "get":
		key := msg.GetObject(tf.conf.Object.Key).String()
		if tf.conf.Prefix != "" {
			key = fmt.Sprint(tf.conf.Prefix, ":", key)
		}

		v, err := tf.kvStore.Get(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("transform: mod_kv_store: %v", err)
		}

		if err := msg.SetObject(tf.conf.Object.SetKey, v); err != nil {
			return nil, fmt.Errorf("transform: mod_kv_store: %v", err)
		}

		return []*message.Message{msg}, nil
	case "set":
		key := msg.GetObject(tf.conf.Object.SetKey).String()
		if tf.conf.Prefix != "" {
			key = fmt.Sprint(tf.conf.Prefix, ":", key)
		}

		//nolint: nestif // ignore nesting complexity
		if tf.conf.TTLKey != "" && tf.conf.TTLOffset != 0 {
			ttl := msg.GetObject(tf.conf.TTLKey).Int() + tf.conf.TTLOffset
			if err := tf.kvStore.SetWithTTL(ctx, key, msg.GetObject(tf.conf.Object.Key).String(), ttl); err != nil {
				return nil, fmt.Errorf("transform: mod_kv_store: %v", err)
			}
		} else if tf.conf.TTLKey != "" {
			ttl := msg.GetObject(tf.conf.TTLKey).Int()
			if err := tf.kvStore.SetWithTTL(ctx, key, msg.GetObject(tf.conf.Object.Key).String(), ttl); err != nil {
				return nil, fmt.Errorf("transform: mod_kv_store: %v", err)
			}
		} else if tf.conf.TTLOffset != 0 {
			ttl := time.Now().Add(time.Duration(tf.conf.TTLOffset) * time.Second).Unix()
			if err := tf.kvStore.SetWithTTL(ctx, key, msg.GetObject(tf.conf.Object.Key).String(), ttl); err != nil {
				return nil, fmt.Errorf("transform: mod_kv_store: %v", err)
			}
		} else {
			if err := tf.kvStore.Set(ctx, key, msg.GetObject(tf.conf.Object.Key).String()); err != nil {
				return nil, fmt.Errorf("transform: mod_kv_store: %v", err)
			}
		}

		return []*message.Message{msg}, nil
	}

	return nil, nil
}
