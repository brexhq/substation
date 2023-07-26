package kv

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/file"
	"github.com/brexhq/substation/internal/json"
)

// kvJSONFile is a read-only key-value store that is derived from a file containing
// an object and stored in memory.
type kvJSONFile struct {
	// File contains the location of the text file. This can be either a path on local
	// disk, an HTTP(S) URL, or an AWS S3 URL.
	File string `json:"file"`
	// IsLines indicates that the file is a JSON Lines file. The first non-null value
	// is returned when a key is found.
	IsLines bool `json:"is_lines"`

	mu     *sync.Mutex
	object []byte
}

// Create a new JSON file KV store.
func newKVJSONFile(cfg config.Config) (*kvJSONFile, error) {
	var store kvJSONFile
	if err := config.Decode(cfg.Settings, &store); err != nil {
		return nil, err
	}
	store.mu = new(sync.Mutex)

	if store.File == "" {
		return nil, fmt.Errorf("kv: json: options %+v: %v", &store, errors.ErrMissingRequiredOption)
	}

	return &store, nil
}

func (store *kvJSONFile) String() string {
	return toString(store)
}

// Get retrieves a value from the store.
func (store *kvJSONFile) Get(ctx context.Context, key string) (interface{}, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// JSON Lines files are queried as an array and the first non-null value is returned.
	// See https://github.com/tidwall/gjson#json-lines for more information.
	if store.IsLines && !strings.HasPrefix(key, "..#.") {
		key = "..#." + key
		res := json.Get(store.object, key)

		for _, v := range res.Array() {
			if json.Types[v.Type] != "Null" {
				return v.Value(), nil
			}
		}
	}

	res := json.Get(store.object, key)
	if json.Types[res.Type] == "Null" {
		return nil, nil
	}

	return res.Value(), nil
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
	store.mu.Lock()
	defer store.mu.Unlock()

	return store.object != nil
}

// Setup creates the store by reading the text file into memory.
func (store *kvJSONFile) Setup(ctx context.Context) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// avoids unnecessary setup
	if store.object != nil {
		return nil
	}

	path, err := file.Get(ctx, store.File)
	defer os.Remove(path)
	if err != nil {
		return fmt.Errorf("kv: json_file: %v", err)
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("kv: json_file: %v", err)
	}

	store.object = buf
	return nil
}

// Closes the store.
func (store *kvJSONFile) Close() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// avoids unnecessary closing
	if store.object == nil {
		return nil
	}

	store.object = nil
	return nil
}
