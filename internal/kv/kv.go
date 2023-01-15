package kv

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

var (
	mu sync.Mutex
	m  map[string]Storer
	// errSetNotSupported is returned when the KV set action is not supported.
	errSetNotSupported = errors.Error("set not supported")
)

// Storer provides tools for getting values from and putting values into key-value stores.
type Storer interface {
	Get(context.Context, string) (interface{}, error)
	Set(context.Context, string, interface{}) error
	SetWithTTL(context.Context, string, interface{}, int64) error
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

	switch t := cfg.Type; t {
	case "aws_dynamodb":
		var c kvAWSDynamoDB
		_ = config.Decode(cfg.Settings, &c)
		m[sig] = &c
	case "csv_file":
		var c kvCSVFile
		_ = config.Decode(cfg.Settings, &c)
		m[sig] = &c
	case "json_file":
		var c kvJSONFile
		_ = config.Decode(cfg.Settings, &c)
		m[sig] = &c
	case "memory":
		var c kvMemory
		_ = config.Decode(cfg.Settings, &c)
		m[sig] = &c
	case "mmdb":
		var c kvMMDB
		_ = config.Decode(cfg.Settings, &c)
		m[sig] = &c
	case "text_file":
		var c kvTextFile
		_ = config.Decode(cfg.Settings, &c)
		m[sig] = &c
	default:
		return nil, fmt.Errorf("kv_store: %s: %v", t, errors.ErrInvalidFactoryInput)
	}

	return m[sig], nil
}

func init() {
	m = make(map[string]Storer)
}
