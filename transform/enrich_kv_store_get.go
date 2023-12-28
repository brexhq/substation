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
	Object iconfig.Object `json:"object"`

	// Prefix is prepended to the key and can be used to simplify
	// data management within a KV store.
	//
	// This is optional and defaults to an empty string.
	Prefix string `json:"prefix"`
	// KVStore determine the type of KV store used by the transform. Refer to internal/kv
	// for more information.
	KVStore config.Config `json:"kv_store"`
	// CloseKVStore determines if the KV store is closed when a control
	// message is received.
	CloseKVStore bool `json:"close_kv_store"`
}

func (c *enrichKVStoreGetConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *enrichKVStoreGetConfig) Validate() error {
	if c.Object.SrcKey == "" {
		return fmt.Errorf("object_src_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.DstKey == "" {
		return fmt.Errorf("object_dst_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.KVStore.Type == "" {
		return fmt.Errorf("kv_store: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newEnrichKVStoreGet(_ context.Context, cfg config.Config) (*enrichKVStoreGet, error) {
	conf := enrichKVStoreGetConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: enrich_kv_store_get: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: enrich_kv_store_get: %v", err)
	}

	kvStore, err := kv.Get(conf.KVStore)
	if err != nil {
		return nil, fmt.Errorf("transform: enrich_kv_store_get: %v", err)
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
			return nil, fmt.Errorf("transform: enrich_kv_store_get: %v", err)
		}

		return []*message.Message{msg}, nil
	}

	if !tf.kvStore.IsEnabled() {
		if err := tf.kvStore.Setup(ctx); err != nil {
			return nil, fmt.Errorf("transform: enrich_kv_store_get: %v", err)
		}
	}

	value := msg.GetValue(tf.conf.Object.SrcKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	key := value.String()
	if tf.conf.Prefix != "" {
		key = fmt.Sprint(tf.conf.Prefix, ":", key)
	}

	v, err := tf.kvStore.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("transform: enrich_kv_store_get: %v", err)
	}

	if err := msg.SetValue(tf.conf.Object.DstKey, v); err != nil {
		return nil, fmt.Errorf("transform: enrich_kv_store_get: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *enrichKVStoreGet) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
