//go:build !wasm

package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/kv"
	"github.com/brexhq/substation/message"
)

type enrichKVStoreGetConfig struct {
	// Prefix is prepended to the key and can be used to simplify
	// data management within a KV store.
	//
	// This is optional and defaults to an empty string.
	Prefix string `json:"prefix"`
	// CloseKVStore determines if the KV store is closed when a control
	// message is received.
	//
	// This is optional and defaults to false (KV store is not closed).
	CloseKVStore bool `json:"close_kv_store"`

	ID      string         `json:"id"`
	Object  iconfig.Object `json:"object"`
	KVStore config.Config  `json:"kv_store"`
}

func (c *enrichKVStoreGetConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *enrichKVStoreGetConfig) Validate() error {
	if c.Object.SourceKey == "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.KVStore.Type == "" {
		return fmt.Errorf("kv_store: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newEnrichKVStoreGet(_ context.Context, cfg config.Config) (*enrichKVStoreGet, error) {
	conf := enrichKVStoreGetConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform enrich_kv_store_get: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "enrich_kv_store_get"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	kvStore, err := kv.Get(conf.KVStore)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := enrichKVStoreGet{
		conf:    conf,
		kvStore: kvStore,
	}

	return &tf, nil
}

type enrichKVStoreGet struct {
	conf    enrichKVStoreGetConfig
	kvStore kv.Storer
}

func (tf *enrichKVStoreGet) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

	v, err := tf.kvStore.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, v); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *enrichKVStoreGet) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
