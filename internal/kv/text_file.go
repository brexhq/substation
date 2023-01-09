package kv

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/brexhq/substation/internal/file"
)

// kvTextFile is a read-only key-value store that is derived from a newline delimited
// text file and stored in memory.
//
// Rows from the text file are stored in a slice where each element becomes the key and
// the value is a boolean true.
//
// For example, if the file contains this data:
//
//	foo
//	bar
//	baz
//
// The store becomes this:
//
//	map[foo:true bar:true baz:true]
type kvTextFile struct {
	// File contains the location of the text file. This can be either a path on local
	// disk, an HTTP(S) URL, or an AWS S3 URL.
	File  string `json:"file"`
	mu    sync.Mutex
	items []string
}

func (store *kvTextFile) String() string {
	return toString(store)
}

// Get retrieves a value from the store.
func (store *kvTextFile) Get(ctx context.Context, key string) (interface{}, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	return store.contains(key), nil
}

// Set is unused because this is a read-only store.
func (store *kvTextFile) Set(ctx context.Context, key string, val interface{}) error {
	return errSetNotSupported
}

// SetWithTTL is unused because this is a read-only store.
func (store *kvTextFile) SetWithTTL(ctx context.Context, key string, val interface{}, ttl int64) error {
	return errSetNotSupported
}

// IsEnabled returns true if the store is ready for use.
func (store *kvTextFile) IsEnabled() bool {
	return store.items != nil
}

// Setup creates the store by reading the text file into memory.
func (store *kvTextFile) Setup(ctx context.Context) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// avoids unnecessary setup
	if store.items != nil {
		return nil
	}

	path, err := file.Get(ctx, store.File)
	defer os.Remove(path)
	if err != nil {
		return fmt.Errorf("kv: text_file: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("kv: text_file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		store.items = append(store.items, scanner.Text())
	}

	return nil
}

// Closes the store.
func (store *kvTextFile) Close() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// avoids unnecessary closing
	if store.items == nil {
		return nil
	}

	store.items = nil
	return nil
}

func (store *kvTextFile) contains(key string) bool {
	for _, item := range store.items {
		if item == key {
			return true
		}
	}

	return false
}
