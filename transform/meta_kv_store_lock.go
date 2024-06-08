package transform

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/kv"
	"github.com/brexhq/substation/message"
)

type metaVStoreLockObjectConfig struct {
	// TTLKey retrieves a value from an object that is used as the time-to-live (TTL)
	// of the item locked in the KV store. This value must be an integer that represents
	// the Unix time when the item will be evicted from the store. Any precision greater
	// than seconds (e.g., milliseconds, nanoseconds) is truncated to seconds.
	//
	// This is optional and defaults to using no TTL when setting items into the store.
	TTLKey string `json:"ttl_key"`

	iconfig.Object
}

type metaKVStoreLockConfig struct {
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

	Transform config.Config              `json:"transform"`
	Object    metaVStoreLockObjectConfig `json:"object"`
	KVStore   config.Config              `json:"kv_store"`
}

func (c *metaKVStoreLockConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaKVStoreLockConfig) Validate() error {
	if c.Transform.Type == "" {
		return fmt.Errorf("transform: %v", errors.ErrMissingRequiredOption)
	}

	if c.KVStore.Type == "" {
		return fmt.Errorf("kv_store: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaKVStoreLock(ctx context.Context, cfg config.Config) (*metaKVStoreLock, error) {
	conf := metaKVStoreLockConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: meta_kv_store_lock: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: meta_kv_store_lock: %v", err)
	}

	tff, err := New(ctx, conf.Transform)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_kv_store_lock: %v", err)
	}

	locker, err := kv.GetLocker(conf.KVStore)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_kv_store_lock: %v", err)
	}

	if err := locker.Setup(ctx); err != nil {
		return nil, fmt.Errorf("transform: meta_kv_store_lock: %v", err)
	}

	if conf.TTLOffset == "" {
		conf.TTLOffset = "0s"
	}

	dur, err := time.ParseDuration(conf.TTLOffset)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_kv_store_lock: %v", err)
	}

	tf := metaKVStoreLock{
		tf:     tff,
		conf:   conf,
		locker: locker,
		ttl:    int64(dur.Seconds()),
	}

	return &tf, nil
}

// metaKVStoreLock applies a lock in a KV store and executes a transform. If the lock is already
// held, then an error is returned. The lock is applied with a time-to-live (TTL) value, which is
// used to determine when the lock is automatically released.
type metaKVStoreLock struct {
	tf     Transformer
	conf   metaKVStoreLockConfig
	locker kv.Locker
	ttl    int64

	// mu is required to prevent concurrent access to the keys slice.
	mu   sync.Mutex
	keys []string
}

// Transforms a message based on the configuration.
func (tf *metaKVStoreLock) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		msgs, err := tf.tf.Transform(ctx, msg)
		if err != nil {
			for _, key := range tf.keys {
				_ = tf.locker.Unlock(ctx, key)
			}

			return nil, fmt.Errorf("transform: meta_kv_store_lock: %v", err)
		}

		tf.keys = tf.keys[:0]
		return msgs, nil
	}

	// By default, the lock key is the SHA256 hash of the message.
	var lockKey string
	v := msg.GetValue(tf.conf.Object.SourceKey)
	if !v.Exists() {
		sum := sha256.Sum256(msg.Data())
		lockKey = fmt.Sprintf("%x", sum)
	} else {
		lockKey = v.String()
	}

	if tf.conf.Prefix != "" {
		lockKey = fmt.Sprint(tf.conf.Prefix, ":", lockKey)
	}

	// Calculate TTL based on the configuration.
	var ttl int64
	if tf.conf.Object.TTLKey != "" && tf.ttl != 0 {
		v := msg.GetValue(tf.conf.Object.TTLKey)
		ttl = truncateTTL(v) + tf.ttl
	} else if tf.conf.Object.TTLKey != "" {
		v := msg.GetValue(tf.conf.Object.TTLKey)
		ttl = truncateTTL(v)
	} else if tf.ttl != 0 {
		ttl = time.Now().Add(time.Duration(tf.ttl) * time.Second).Unix()
	}

	// Acquire the lock. If the lock is already held, then the message is returned as is.
	// This prevents the transform from being applied to the message more than once.
	if err := tf.locker.Lock(ctx, lockKey, ttl); err != nil {
		if err == kv.ErrNoLock {
			return []*message.Message{msg}, nil
		}

		for _, key := range tf.keys {
			_ = tf.locker.Unlock(ctx, key)
		}

		return nil, fmt.Errorf("transform: meta_kv_store_lock: %v", err)
	}

	tf.keys = append(tf.keys, lockKey)
	msgs, err := tf.tf.Transform(ctx, msg)
	if err != nil {
		for _, key := range tf.keys {
			_ = tf.locker.Unlock(ctx, key)
		}

		return nil, fmt.Errorf("transform: meta_kv_store_lock: %v", err)
	}

	return msgs, nil
}
