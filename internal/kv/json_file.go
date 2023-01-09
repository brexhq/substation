package kv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/brexhq/substation/internal/file"
)

// kvJSONFile is a read-only key-value store that is derived from a file containing
// an object and stored in memory.
//
// For example, if the file contains this data:
//
//	{"foo":"bar","baz":"qux","quux":"corge"}
//
// The store becomes this:
//
//	map[foo:bar baz:qux quux:corge]
type kvJSONFile struct {
	// File contains the location of the text file. This can be either a path on local
	// disk, an HTTP(S) URL, or an AWS S3 URL.
	File  string `json:"file"`
	mu    sync.Mutex
	items map[string]interface{}
}

func (store *kvJSONFile) String() string {
	return toString(store)
}

// Get retrieves a value from the store.
func (store *kvJSONFile) Get(ctx context.Context, key string) (interface{}, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if val, ok := store.items[key]; ok {
		return val, nil
	}

	return nil, nil
}

// Set is unused because this is a read-only store.
func (store *kvJSONFile) Set(ctx context.Context, key string, val interface{}) error {
	return errSetNotSupported
}

// SetWithTTL is unused because this is a read-only store.
func (store *kvJSONFile) SetWithTTL(ctx context.Context, key string, val interface{}, ttl int64) error {
	return errSetNotSupported
}

// IsEnabled returns true if the store is ready for use.
func (store *kvJSONFile) IsEnabled() bool {
	return store.items != nil
}

// Setup creates the store by reading the text file into memory.
func (store *kvJSONFile) Setup(ctx context.Context) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// avoids unnecessary setup
	if store.items != nil {
		return nil
	}

	store.items = make(map[string]interface{})

	path, err := file.Get(ctx, store.File)
	defer os.Remove(path)
	if err != nil {
		return fmt.Errorf("kv: json_file: %v", err)
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("kv: json_file: %v", err)
	}

	if err := json.Unmarshal(buf, &store.items); err != nil {
		return fmt.Errorf("kv: json_file: %v", err)
	}

	return nil
}

// Closes the store.
func (store *kvJSONFile) Close() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// avoids unnecessary closing
	if store.items == nil {
		return nil
	}

	store.items = nil
	return nil
}
