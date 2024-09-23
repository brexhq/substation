package transform

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/kv"
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
	// Transforms that are applied in series after the lock is acquired.
	Transforms []config.Config `json:"transforms"`

	ID      string                     `json:"id"`
	Object  metaVStoreLockObjectConfig `json:"object"`
	KVStore config.Config              `json:"kv_store"`
}

func (c *metaKVStoreLockConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaKVStoreLockConfig) Validate() error {
	if len(c.Transforms) == 0 {
		return fmt.Errorf("transform: %v", iconfig.ErrMissingRequiredOption)
	}

	for _, t := range c.Transforms {
		if t.Type == "" {
			return fmt.Errorf("transform: %v", iconfig.ErrMissingRequiredOption)
		}
	}

	if c.KVStore.Type == "" {
		return fmt.Errorf("kv_store: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaKVStoreLock(ctx context.Context, cfg config.Config) (*metaKVStoreLock, error) {
	conf := metaKVStoreLockConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform meta_kv_store_lock: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "meta_kv_store_lock"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := metaKVStoreLock{
		conf: conf,
	}

	tf.tfs = make([]Transformer, len(conf.Transforms))
	for i, t := range conf.Transforms {
		tfer, err := New(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
		}

		tf.tfs[i] = tfer
	}

	locker, err := kv.GetLocker(conf.KVStore)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	if err := locker.Setup(ctx); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}
	tf.locker = locker

	if conf.TTLOffset == "" {
		conf.TTLOffset = "0s"
	}

	dur, err := time.ParseDuration(conf.TTLOffset)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}
	tf.ttl = int64(dur.Seconds())

	return &tf, nil
}

// metaKVStoreLock applies a lock in a KV store and executes a transform. If the lock is already
// held, then an error is returned. The lock is applied with a time-to-live (TTL) value, which is
// used to determine when the lock is automatically released.
type metaKVStoreLock struct {
	tfs []Transformer

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
		msgs, err := Apply(ctx, tf.tfs, msg)
		if err != nil {
			for _, key := range tf.keys {
				_ = tf.locker.Unlock(ctx, key)
			}

			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
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

		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	tf.keys = append(tf.keys, lockKey)

	msgs, err := Apply(ctx, tf.tfs, msg)
	if err != nil {
		for _, key := range tf.keys {
			_ = tf.locker.Unlock(ctx, key)
		}

		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return msgs, nil
}
