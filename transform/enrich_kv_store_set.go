//go:build !wasm

package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/kv"
	"github.com/brexhq/substation/message"
)

type enrichKVStoreSetConfig struct {
	Object iconfig.Object `json:"object"`

	// Prefix is prepended to the key and can be used to simplify
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

func (c *enrichKVStoreSetConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *enrichKVStoreSetConfig) Validate() error {
	if c.Object.Key == "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.KVStore.Type == "" {
		return fmt.Errorf("kv_store: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newEnrichKVStoreSet(_ context.Context, cfg config.Config) (*enrichKVStoreSet, error) {
	conf := enrichKVStoreSetConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: enrich_kv_store_set: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: enrich_kv_store_set: %v", err)
	}

	kvStore, err := kv.Get(conf.KVStore)
	if err != nil {
		return nil, fmt.Errorf("transform: enrich_kv_store_set: %v", err)
	}

	tf := enrichKVStoreSet{
		conf:    conf,
		kvStore: kvStore,
	}

	return &tf, nil
}

type enrichKVStoreSet struct {
	conf    enrichKVStoreSetConfig
	kvStore kv.Storer
}

func (tf *enrichKVStoreSet) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		if !tf.conf.CloseKVStore {
			return []*message.Message{msg}, nil
		}

		if err := tf.kvStore.Close(); err != nil {
			return nil, fmt.Errorf("transform: enrich_kv_store_set: %v", err)
		}

		return []*message.Message{msg}, nil
	}

	if !tf.kvStore.IsEnabled() {
		if err := tf.kvStore.Setup(ctx); err != nil {
			return nil, fmt.Errorf("transform: enrich_kv_store_set: %v", err)
		}
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	key := value.String()
	if tf.conf.Prefix != "" {
		key = fmt.Sprint(tf.conf.Prefix, ":", key)
	}

	//nolint: nestif // ignore nesting complexity
	if tf.conf.TTLKey != "" && tf.conf.TTLOffset != 0 {
		ttl := msg.GetValue(tf.conf.TTLKey).Int() + tf.conf.TTLOffset
		if err := tf.kvStore.SetWithTTL(ctx, key, msg.GetValue(tf.conf.Object.SetKey).String(), ttl); err != nil {
			return nil, fmt.Errorf("transform: enrich_kv_store_set: %v", err)
		}
	} else if tf.conf.TTLKey != "" {
		ttl := msg.GetValue(tf.conf.TTLKey).Int()
		if err := tf.kvStore.SetWithTTL(ctx, key, msg.GetValue(tf.conf.Object.SetKey).String(), ttl); err != nil {
			return nil, fmt.Errorf("transform: enrich_kv_store_set: %v", err)
		}
	} else if tf.conf.TTLOffset != 0 {
		ttl := time.Now().Add(time.Duration(tf.conf.TTLOffset) * time.Second).Unix()
		if err := tf.kvStore.SetWithTTL(ctx, key, msg.GetValue(tf.conf.Object.SetKey).String(), ttl); err != nil {
			return nil, fmt.Errorf("transform: enrich_kv_store_set: %v", err)
		}
	} else {
		if err := tf.kvStore.Set(ctx, key, msg.GetValue(tf.conf.Object.SetKey).String()); err != nil {
			return nil, fmt.Errorf("transform: enrich_kv_store_set: %v", err)
		}
	}

	return []*message.Message{msg}, nil
}

func (tf *enrichKVStoreSet) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
