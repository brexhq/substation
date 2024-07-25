package kv

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
)

// kvMemory is a read-write key-value store that is stored in memory.
//
// This KV store uses least recently used (LRU) eviction and optionally supports
// per-value time-to-live (TTL).
type kvMemory struct {
	// Capacity limits the maximum capacity of the store.
	//
	// This is optional and defaults to 1024 values.
	Capacity int `json:"capacity"`
	mu       sync.Mutex
	lockMu   sync.Mutex
	lru      list.List
	items    map[string]*list.Element
}

// Create a new memory KV store.
func newKVMemory(cfg config.Config) (*kvMemory, error) {
	var store kvMemory
	if err := _config.Decode(cfg.Settings, &store); err != nil {
		return nil, err
	}

	return &store, nil
}

func (store *kvMemory) String() string {
	return toString(store)
}

type kvMemoryElement struct {
	key   string
	value interface{}
	ttl   int64
}

// Get retrieves a value from the store. If the value had a time-to-live (TTL)
// configured when it was added and the TTL has passed, then nothing is returned.
func (store *kvMemory) Get(ctx context.Context, key string) (interface{}, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if node, found := store.items[key]; found {
		ttl := node.Value.(kvMemoryElement).ttl

		// a zero value for ttl indicates that ttl is not configured for the item
		if ttl != 0 && ttl <= time.Now().Unix() {
			delete(store.items, key)
			store.lru.Remove(node)

			return nil, nil
		}

		// resetting the position of the node prevents recently accessed items from being evicted
		store.lru.MoveToFront(node)
		return node.Value.(kvMemoryElement).value, nil
	}

	return nil, nil
}

// Set adds a value to the store. If the addition causes the capacity of the store to
// exceed the configured limit, then the least recently accessed value is removed from
// the store.
func (store *kvMemory) Set(ctx context.Context, key string, val interface{}) error {
	return store.SetWithTTL(ctx, key, val, 0)
}

// SetWithTTL adds a value to the store with a time-to-live (TTL). If the addition
// causes the capacity of the store to exceed the configured limit, then the least
// recently accessed value is removed from the store.
func (store *kvMemory) SetWithTTL(ctx context.Context, key string, val interface{}, ttl int64) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	value := kvMemoryElement{key, val, ttl}

	if node, ok := store.items[key]; ok {
		// resetting the position of the node prevents recently accessed items from being evicted
		store.lru.MoveToFront(node)
		node.Value = value

		return nil
	}

	store.lru.PushFront(value)
	store.items[key] = store.lru.Front()

	if store.lru.Len() > store.Capacity {
		node := store.lru.Back()

		store.lru.Remove(node)
		delete(store.items, node.Value.(kvMemoryElement).key)
	}

	return nil
}

// Lock adds an item to the store if it does not already exist. If the item already exists
// and the time-to-live (TTL) has not expired, then this returns ErrNoLock.
func (store *kvMemory) Lock(ctx context.Context, key string, ttl int64) error {
	store.lockMu.Lock()
	defer store.lockMu.Unlock()

	if node, ok := store.items[key]; ok {
		ttl := node.Value.(kvMemoryElement).ttl
		if ttl <= time.Now().Unix() {
			delete(store.items, key)
			store.lru.Remove(node)
		}

		return ErrNoLock
	}

	return store.SetWithTTL(ctx, key, nil, ttl)
}

// AppendWithTTL appends a value to a list in the store. If the list does not exist, then
// it is created. If a non-zero TTL is provided, then the TTL value is also updated.
func (store *kvMemory) AppendWithTTL(ctx context.Context, key string, val interface{}, ttl int64) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// List already exists for the key.
	if node, ok := store.items[key]; ok {
		// Resetting the position of the node prevents recently accessed items from being evicted
		store.lru.MoveToFront(node)

		node.Value = kvMemoryElement{
			key:   key,
			value: append(node.Value.(kvMemoryElement).value.([]interface{}), val),
			ttl:   ttl, // Always update the TTL value. Zero values are ignored on retrieval.
		}

		return nil
	}

	// No list exists for the key.
	store.lru.PushFront(kvMemoryElement{key, []interface{}{val}, ttl})
	store.items[key] = store.lru.Front()

	if store.lru.Len() > store.Capacity {
		node := store.lru.Back()

		store.lru.Remove(node)
		delete(store.items, node.Value.(kvMemoryElement).key)
	}

	return nil
}

// Unlock removes an item from the store.
func (store *kvMemory) Unlock(ctx context.Context, key string) error {
	store.lockMu.Lock()
	defer store.lockMu.Unlock()

	if node, ok := store.items[key]; ok {
		store.lru.Remove(node)
		delete(store.items, key)
	}

	return nil
}

// IsEnabled returns true if the store is ready for use.
func (store *kvMemory) IsEnabled() bool {
	store.mu.Lock()
	defer store.mu.Unlock()

	return store.items != nil
}

// Setup creates the store.
func (store *kvMemory) Setup(ctx context.Context) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// avoids unnecessary setup
	if store.items != nil {
		return nil
	}

	store.items = make(map[string]*list.Element)

	if store.Capacity == 0 {
		store.Capacity = 1024
	}

	return nil
}

// Closes the store.
func (store *kvMemory) Close() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// avoids unnecessary closing
	if store.items == nil {
		return nil
	}

	store.items = nil
	return nil
}
