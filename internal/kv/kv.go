package kv

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/v2/config"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

var (
	mu   sync.Mutex
	m    map[string]Storer
	lock map[string]Locker
	// errSetNotSupported is returned when the KV set action is not supported.
	errSetNotSupported = fmt.Errorf("set not supported")
	// ErrNoLock is returned when a lock cannot be acquired.
	ErrNoLock = fmt.Errorf("unable to acquire lock")
)

// Storer provides tools for getting values from and putting values into key-value stores.
type Storer interface {
	Get(context.Context, string) (interface{}, error)
	Set(context.Context, string, interface{}) error
	SetWithTTL(context.Context, string, interface{}, int64) error
	SetAddWithTTL(context.Context, string, interface{}, int64) error
	Setup(context.Context) error
	Close() error
	IsEnabled() bool
}

// required to support Stringer interface
func toString(s Storer) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// Get returns a pointer to a Storer that is stored as a package level global variable.
// This function and each Storer are safe for concurrent access.
func Get(cfg config.Config) (Storer, error) {
	mu.Lock()
	defer mu.Unlock()

	// KV store configurations are mapped using the "signature" of their config.
	// this makes it possible for a single run of a Substation application to rely
	// on multiple KV stores.
	sig := fmt.Sprint(cfg)
	store, ok := m[sig]
	if ok {
		return store, nil
	}

	storer, err := New(cfg)
	if err != nil {
		return nil, err
	}
	m[sig] = storer

	return m[sig], nil
}

// New returns a Storer.
func New(cfg config.Config) (Storer, error) {
	switch t := cfg.Type; t {
	case "aws_dynamodb":
		return newKVAWSDynamoDB(cfg)
	case "csv_file":
		return newKVCSVFile(cfg)
	case "json_file":
		return newKVJSONFile(cfg)
	case "memory":
		return newKVMemory(cfg)
	case "mmdb":
		return newKVMMDB(cfg)
	case "text_file":
		return newKVTextFile(cfg)
	default:
		return nil, fmt.Errorf("kv_store: %s: %v", t, iconfig.ErrInvalidFactoryInput)
	}
}

type Locker interface {
	Lock(context.Context, string, int64) error
	Unlock(context.Context, string) error
	Setup(context.Context) error
	IsEnabled() bool
}

// Get returns a pointer to a Locker that is stored as a package level global variable.
// This function and each Locker are safe for concurrent access.
func GetLocker(cfg config.Config) (Locker, error) {
	mu.Lock()
	defer mu.Unlock()

	// KV store configurations are mapped using the "signature" of their config.
	// this makes it possible for a single run of a Substation application to rely
	// on multiple KV stores.
	sig := fmt.Sprint(cfg)
	locker, ok := lock[sig]
	if ok {
		return locker, nil
	}

	locker, err := NewLocker(cfg)
	if err != nil {
		return nil, err
	}
	lock[sig] = locker

	return lock[sig], nil
}

func NewLocker(cfg config.Config) (Locker, error) {
	switch t := cfg.Type; t {
	case "aws_dynamodb":
		return newKVAWSDynamoDB(cfg)
	case "memory":
		return newKVMemory(cfg)
	default:
		return nil, fmt.Errorf("kv_store locker: %s: %v", t, iconfig.ErrInvalidFactoryInput)
	}
}

func init() {
	m = make(map[string]Storer)
	lock = make(map[string]Locker)
}
