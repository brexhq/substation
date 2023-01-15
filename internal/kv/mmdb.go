package kv

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/file"
	"github.com/oschwald/maxminddb-golang"
)

// errMMDBKeyMustBeAddr is returned when the key used in a Get call is not a valid
// IP address.
var errMMDBKeyMustBeAddr = errors.Error("key must be IP address")

// KvMMDB is a read-only key-value store that is derived from any MaxMind database
// format (MMDB) file.
//
// MMDB is an open source database file format that maps IPv4 and IPv6 addresses to
// data records, and is most commonly utilized by MaxMind GeoIP databases. Learn more
// about the file format here: https://maxmind.github.io/MaxMind-DB/.
type kvMMDB struct {
	// File contains the location of the MMDB file. This can be either a path on local
	// disk, an HTTP(S) URL, or an AWS S3 URL.
	File   string `json:"file"`
	mu     sync.Mutex
	reader *maxminddb.Reader
}

func (store *kvMMDB) String() string {
	return toString(store)
}

// Get retrieves a value from the store.
func (store *kvMMDB) Get(ctx context.Context, key string) (interface{}, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	addr := net.ParseIP(key)
	if addr == nil {
		// does not include the key that caused the error to avoid leaking
		// private data. this should be wrapped by the caller, which can
		// provide more information about what caused the error.
		return nil, fmt.Errorf("kv: mmdb: %v", errMMDBKeyMustBeAddr)
	}

	var r interface{}
	if err := store.reader.Lookup(addr, &r); err != nil {
		return nil, fmt.Errorf("kv: mmdb: %v", err)
	}

	return r, nil
}

// Set is unused because this is a read-only store.
func (store *kvMMDB) Set(ctx context.Context, key string, val interface{}) error {
	return errSetNotSupported
}

// SetWithTTL is unused because this is a read-only store.
func (store *kvMMDB) SetWithTTL(ctx context.Context, key string, val interface{}, ttl int64) error {
	return errSetNotSupported
}

// IsEnabled returns true if the store is ready for use.
func (store *kvMMDB) IsEnabled() bool {
	store.mu.Lock()
	defer store.mu.Unlock()

	return store.reader != nil
}

// Setup creates the store by reading the text file into memory.
func (store *kvMMDB) Setup(ctx context.Context) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// avoids unnecessary setup
	if store.reader != nil {
		return nil
	}

	path, err := file.Get(ctx, store.File)
	defer os.Remove(path)
	if err != nil {
		return fmt.Errorf("kv: mmdb: %v", err)
	}

	db, err := maxminddb.Open(path)
	if err != nil {
		return fmt.Errorf("kv: mmdb: %v", err)
	}

	store.reader = db
	return nil
}

// Closes the store.
func (store *kvMMDB) Close() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// avoids unnecessary closing
	if store.reader == nil {
		return nil
	}

	if err := store.reader.Close(); err != nil {
		return fmt.Errorf("kv: mmdb: %v", err)
	}

	return nil
}
