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
	// errMissingRequiredOptions is returned when a KV store does not have the required options to properly execute.
	errMissingRequiredOptions = errors.Error("missing required options")
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
		c, err := newKVAWSDyanmoDB(cfg)
		if err != nil {
			return nil, err
		}
		m[sig] = &c
	case "csv_file":
		c, err := newKVCSVFile(cfg)
		if err != nil {
			return nil, err
		}
		m[sig] = &c
	case "json_file":
		c, err := newKVJSONFile(cfg)
		if err != nil {
			return nil, err
		}
		m[sig] = &c
	case "memory":
		c, err := newKVMemory(cfg)
		if err != nil {
			return nil, err
		}
		m[sig] = &c
	case "mmdb":
		c, err := newKVMMDB(cfg)
		if err != nil {
			return nil, err
		}
		m[sig] = &c
	case "text_file":
		c, err := newKVTextFile(cfg)
		if err != nil {
			return nil, err
		}
		m[sig] = &c
	default:
		return nil, fmt.Errorf("kv_store: %s: %v", t, errors.ErrInvalidFactoryInput)
	}

	return m[sig], nil
}

// New returns a Storer.
func New(cfg config.Config) (Storer, error) {
	switch t := cfg.Type; t {
	case "aws_dynamodb":
		c, err := newKVAWSDyanmoDB(cfg)
		return &c, err
	case "csv_file":
		c, err := newKVCSVFile(cfg)
		return &c, err
	case "json_file":
		c, err := newKVJSONFile(cfg)
		return &c, err
	case "memory":
		c, err := newKVMemory(cfg)
		return &c, err
	case "mmdb":
		c, err := newKVMMDB(cfg)
		return &c, err
	case "text_file":
		c, err := newKVTextFile(cfg)
		return &c, err
	default:
		return nil, fmt.Errorf("kv_store: %s: %v", t, errors.ErrInvalidFactoryInput)
	}
}

func init() {
	m = make(map[string]Storer)
}
