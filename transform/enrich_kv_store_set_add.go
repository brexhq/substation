package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/kv"
)

type enrichKVStoreSetAddObjectConfig struct {
	// TTLKey retrieves a value from an object that is used as the time-to-live (TTL)
	// of the item set into the KV store. This value must be an integer that represents
	// the Unix time when the item will be evicted from the store. Any precision greater
	// than seconds (e.g., milliseconds, nanoseconds) is truncated to seconds.
	//
	// This is optional and defaults to using no TTL when setting items into the store.
	TTLKey string `json:"ttl_key"`

	iconfig.Object
}

type enrichKVStoreSetAddConfig struct {
	// Prefix is prepended to the key and can be used to simplify
	// data management within a KV store.
	//
	// This is optional and defaults to an empty string.
	Prefix string `json:"prefix"`
	// TTLOffset is an offset used to determine the time-to-live (TTL) of the item set
	// into the KV store. If Object.TTLKey is configured, then this value is added to the TTL
	// value retrieved from the object. If Object.TTLKey is not used, then this value is added
	// to the current time.
	//
	// For example, if Object.TTLKey is not set and the offset is "1d", then the value
	// will be evicted from the store when more than 1 day has passed.
	//
	// This is optional and defaults to using no TTL when setting values into the store.
	TTLOffset string `json:"ttl_offset"`
	// CloseKVStore determines if the KV store is closed when a control
	// message is received.
	//
	// This is optional and defaults to false (KV store is not closed).
	CloseKVStore bool `json:"close_kv_store"`

	ID      string                          `json:"id"`
	Object  enrichKVStoreSetAddObjectConfig `json:"object"`
	KVStore config.Config                   `json:"kv_store"`
}

func (c *enrichKVStoreSetAddConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *enrichKVStoreSetAddConfig) Validate() error {
	if c.Object.SourceKey == "" {
		return fmt.Errorf("object_source_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.KVStore.Type == "" {
		return fmt.Errorf("kv_store: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func newEnrichKVStoreSetAdd(_ context.Context, cfg config.Config) (*enrichKVStoreSetAdd, error) {
	conf := enrichKVStoreSetAddConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform enrich_kv_store_set: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "enrich_kv_store_set"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	kvStore, err := kv.Get(conf.KVStore)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	if conf.TTLOffset == "" {
		conf.TTLOffset = "0s"
	}

	dur, err := time.ParseDuration(conf.TTLOffset)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := enrichKVStoreSetAdd{
		conf:    conf,
		kvStore: kvStore,
		ttl:     int64(dur.Seconds()),
	}

	return &tf, nil
}

type enrichKVStoreSetAdd struct {
	conf    enrichKVStoreSetAddConfig
	kvStore kv.Storer
	ttl     int64
}

func (tf *enrichKVStoreSetAdd) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		if !tf.conf.CloseKVStore {
			return []*message.Message{msg}, nil
		}

		if err := tf.kvStore.Close(); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		return []*message.Message{msg}, nil
	}

	if !tf.kvStore.IsEnabled() {
		if err := tf.kvStore.Setup(ctx); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	key := value.String()
	if tf.conf.Prefix != "" {
		key = fmt.Sprint(tf.conf.Prefix, ":", key)
	}

	//nolint: nestif // ignore nesting complexity
	if tf.conf.Object.TTLKey != "" && tf.ttl != 0 {
		value := msg.GetValue(tf.conf.Object.TTLKey)
		ttl := truncateTTL(value) + tf.ttl

		if err := tf.kvStore.SetAddWithTTL(ctx, key, msg.GetValue(tf.conf.Object.TargetKey).Value(), ttl); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	} else if tf.conf.Object.TTLKey != "" {
		value := msg.GetValue(tf.conf.Object.TTLKey)
		ttl := truncateTTL(value)

		if err := tf.kvStore.SetAddWithTTL(ctx, key, msg.GetValue(tf.conf.Object.TargetKey).Value(), ttl); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	} else if tf.ttl != 0 {
		ttl := time.Now().Add(time.Duration(tf.ttl) * time.Second).Unix()

		if err := tf.kvStore.SetAddWithTTL(ctx, key, msg.GetValue(tf.conf.Object.TargetKey).Value(), ttl); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	} else {
		if err := tf.kvStore.SetAddWithTTL(ctx, key, msg.GetValue(tf.conf.Object.TargetKey).Value(), 0); err != nil {
			return nil, fmt.Errorf("transform	%s: %v", tf.conf.ID, err)
		}
	}

	return []*message.Message{msg}, nil
}

func (tf *enrichKVStoreSetAdd) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
