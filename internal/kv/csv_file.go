package kv

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/file"
)

// errCSVFileColumnNotFound is returned when the column is not found in the CSV header.
var errCSVFileColumnNotFound = fmt.Errorf("column not found")

// kvCSVFile is a read-only key-value store that is derived from a CSV file and
// stored in memory.
//
// Rows from the CSV are identified by column and stored in a map where the value
// from the column becomes the key and the remaining values from the row become the
// value. Values in the store are string maps of interfaces that can be marshaled to
// an object.
//
// For example, if the file contains this data:
//
//	foo,bar,baz
//	qux,quux,corge
//	grault,garply,waldo
//	fred,plugh,xyzzy
//
// By setting the column to "bar", the store becomes this:
//
//	map[garply:map[baz:waldo foo:grault] plugh:map[baz:xyzzy foo:fred] quux:map[baz:corge foo:qux]]
//
// If the key "garply" is accessed, then values from the store can be marshaled to objects:
//
//	{"baz":"waldo","foo":"grault"}
type kvCSVFile struct {
	// File contains the location of the CSV file. This can be either a path on local
	// disk, an HTTP(S) URL, or an AWS S3 URL.
	File string `json:"file"`
	// Column determines which rows from the CSV file are loaded into the store as keys.
	Column string `json:"column"`
	// Delimiter is the delimiting character (e.g., comma, tab) that separates values
	// in rows in the CSV file.
	//
	// This is optional and defaults to comma (",").
	Delimiter string `json:"delimiter"`
	// Header overrides the header in the CSV file.
	//
	// This is optional and defaults to using the first line of the CSV file as the
	// header.
	Header string `json:"header"`
	mu     sync.Mutex
	items  map[string]map[string]interface{}
}

// Create a new CSV file KV store.
func newKVCSVFile(cfg config.Config) (*kvCSVFile, error) {
	var store kvCSVFile
	if err := _config.Decode(cfg.Settings, &store); err != nil {
		return nil, err
	}

	if store.File == "" || store.Column == "" {
		return nil, fmt.Errorf("kv: csv: options %+v: %v", &store, errors.ErrMissingRequiredOption)
	}

	return &store, nil
}

func (store *kvCSVFile) String() string {
	return toString(store)
}

// Get retrieves a value from the store.
func (store *kvCSVFile) Get(ctx context.Context, key string) (interface{}, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if val, ok := store.items[key]; ok {
		return val, nil
	}

	return nil, nil
}

// Set is unused because this is a read-only store.
func (store *kvCSVFile) Set(ctx context.Context, key string, val interface{}) error {
	return errSetNotSupported
}

// SetWithTTL is unused because this is a read-only store.
func (store *kvCSVFile) SetWithTTL(ctx context.Context, key string, val interface{}, ttl int64) error {
	return errSetNotSupported
}

// IsEnabled returns true if the store is ready for use.
func (store *kvCSVFile) IsEnabled() bool {
	store.mu.Lock()
	defer store.mu.Unlock()

	return store.items != nil
}

// Setup creates the store by reading the CSV file into memory.
func (store *kvCSVFile) Setup(ctx context.Context) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// avoids unnecessary setup
	if store.items != nil {
		return nil
	}

	store.items = make(map[string]map[string]interface{})

	path, err := file.Get(ctx, store.File)
	defer os.Remove(path)
	if err != nil {
		return fmt.Errorf("kv: csv_file: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("kv: csv_file: %v", err)
	}

	defer f.Close()

	var reader *csv.Reader
	// the first line of the CSV file is replaced with header if it exists
	if store.Header != "" {
		buf, err := bufio.NewReader(f).ReadString('\n')
		if err != nil {
			return fmt.Errorf("kv: csv_file: %v", err)
		}
		if _, err = f.Seek(int64(len(buf)), io.SeekStart); err != nil {
			return fmt.Errorf("kv: csv_file: %v", err)
		}

		h := strings.NewReader(fmt.Sprintf("%s\n", store.Header))
		reader = csv.NewReader(io.MultiReader(h, f))
	} else {
		reader = csv.NewReader(f)
	}

	if store.Delimiter == "" {
		store.Delimiter = ","
	}

	// CSV reader only accepts runes for the comma / delimiter
	r, _ := utf8.DecodeRune([]byte(store.Delimiter))
	reader.Comma = r

	// any errors in the CSV file are raised here
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("kv: csv_file: %v", err)
	}

	var header []string
	for i, row := range rows {
		if i == 0 {
			for i := 0; i < len(row); i++ {
				header = append(header, row[i])
			}
		} else {
			// the KV store key is the column's value
			var key string
			for i := 0; i < len(row); i++ {
				if header[i] != store.Column {
					continue
				}

				key = row[i]
			}

			if key == "" {
				return fmt.Errorf("kv: csv_file: %v", errCSVFileColumnNotFound)
			}

			// the KV store value is the row with the column's value removed
			val := make(map[string]interface{})
			for i := 0; i < len(row); i++ {
				if header[i] == store.Column {
					continue
				}

				val[header[i]] = row[i]
			}

			store.items[key] = val
		}
	}

	return nil
}

// Closes the store.
func (store *kvCSVFile) Close() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// avoids unnecessary closing
	if store.items == nil {
		return nil
	}

	store.items = nil
	return nil
}
