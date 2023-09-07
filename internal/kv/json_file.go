package kv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/file"
	"github.com/tidwall/gjson"
)

// errJSONFileInvalid is returned when the file contains invalid JSON.
var errJSONFileInvalid = fmt.Errorf("invalid JSON")

// kvJSONFile is a read-only key-value store that is derived from a file containing
// an object and stored in memory.
type kvJSONFile struct {
	// File contains the location of the text file. This can be either a path on local
	// disk, an HTTP(S) URL, or an AWS S3 URL.
	File   string `json:"file"`
	mu     *sync.Mutex
	object []byte
}

// Create a new JSON file KV store.
func newKVJSONFile(cfg config.Config) (*kvJSONFile, error) {
	var store kvJSONFile
	if err := _config.Decode(cfg.Settings, &store); err != nil {
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

	res := gjson.GetBytes(store.object, key)
	if !res.Exists() {
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

	if !json.Valid(buf) {
		return fmt.Errorf("kv: json_file: %v", errJSONFileInvalid)
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
