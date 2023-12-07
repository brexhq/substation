package transform

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/kv"
	"github.com/brexhq/substation/message"
)

type metaKVStoreLockConfig struct {
	Transform config.Config `json:"transform"`
	LockKey   string        `json:"lock_key"`
	// Prefix is prepended to the key and can be used to simplify
	// data management within a KV store.
	//
	// This is optional and defaults to an empty string.
	Prefix string `json:"prefix"`
	// TTLKey retrieves a value from an object that is used as the time-to-live (TTL)
	// of the item set into the KV store. This value must be an integer that represents
	// the Unix time when the item will be evicted from the store. Any precision greater
	// than seconds (e.g., milliseconds, nanoseconds) is truncated to seconds.
	//
	// This is optional and defaults to using no TTL when setting items into the store.
	TTLKey string `json:"ttl_key"`
	// TTLOffset is an offset used to determine the time-to-live (TTL) of the item set
	// into the KV store. If TTLKey is configured, then this value is added to the TTL
	// value retrieved from the object. If TTLKey is not used, then this value is added
	// to the current time.
	//
	// For example, if TTLKey is not set and the offset is "1d", then the value
	// will be evicted from the store when more than 1 day has passed.
	//
	// This is optional and defaults to using no TTL when setting values into the store.
	TTLOffset string `json:"ttl_offset"`
	// KVStore determine the type of KV store used by the transform. Refer to internal/kv
	// for more information.
	KVStore config.Config `json:"kv_store"`
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

	// Both of these cannot be empty; if they are, then no TTL value exists.
	if c.TTLKey == "" && c.TTLOffset == "" {
		return fmt.Errorf("ttl: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaKVStoreLock(ctx context.Context, cfg config.Config) (*metaKVStoreLock, error) {
	conf := metaKVStoreLockConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: meta_idempotency: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: meta_idempotency: %v", err)
	}

	tff, err := New(ctx, conf.Transform)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_idempotency: %v", err)
	}

	locker, err := kv.NewLocker(conf.KVStore)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_idempotency: %v", err)
	}

	dur, err := time.ParseDuration(conf.TTLOffset)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_idempotency: %v", err)
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
// held, then the message is returned with no transformation applied. The lock is applied with a
// time-to-live (TTL) value, which is used to determine when the lock is automatically released.
// This transform is experimental and may be changed in the future.
//
// TODO:
//
//   - Unlock the KV store if an error occurs during transformation.
//
//   - Add support for updating the KV store with the transformed value after the lock is acquired.
//
//   - Add support for retrieving the value from the KV store and updating the message with the value.
type metaKVStoreLock struct {
	tf     Transformer
	conf   metaKVStoreLockConfig
	locker kv.Locker
	ttl    int64
}

// Transforms a message based on the configuration.
func (tf *metaKVStoreLock) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.locker.IsEnabled() {
		if err := tf.locker.Setup(ctx); err != nil {
			return nil, fmt.Errorf("transform: meta_idempotency: %v", err)
		}
	}

	// By default, the lock key is the SHA256 hash of the message.
	var lockKey string
	v := msg.GetValue(tf.conf.LockKey)
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
	if tf.conf.TTLKey != "" && tf.ttl != 0 {
		v := msg.GetValue(tf.conf.TTLKey)
		ttl = truncateTTL(v) + tf.ttl
	} else if tf.conf.TTLKey != "" {
		v := msg.GetValue(tf.conf.TTLKey)
		ttl = truncateTTL(v)
	} else if tf.ttl != 0 {
		ttl = time.Now().Add(time.Duration(tf.ttl) * time.Second).Unix()
	}

	// Acquire the lock. If the lock is already held, then the message is returned. This
	// prevents the transform from being applied to the message more than once.
	//
	// TODO: If the lock is held, then optionally retrieve the value from the KV store and
	// update the message with the value.
	if err := tf.locker.Lock(ctx, lockKey, ttl); err != nil {
		if err == kv.ErrNoLock {
			return []*message.Message{msg}, nil
		} else {
			return nil, fmt.Errorf("transform: meta_idempotency: %v", err)
		}
	}

	msgs, err := tf.tf.Transform(ctx, msg)
	if err != nil {
		// TODO: Release the lock if the transform fails.
		return nil, fmt.Errorf("transform: meta_idempotency: %v", err)
	}

	// TODO: Update the KV store with a string representation of the transformed value.

	return msgs, nil
}
