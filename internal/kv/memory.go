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
