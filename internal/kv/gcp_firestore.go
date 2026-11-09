package kv

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/brexhq/substation/v2/config"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

// errInvalidKey is returned when a key cannot be used as a Firestore document ID.
var errInvalidKey = fmt.Errorf("invalid key")

// kvGCPFirestore is a read-write key-value store that is backed by a Google Cloud
// Firestore collection.
//
// Each key is stored as a document (keyed by document ID) within the configured
// collection. Values are read from and written to a configurable document field,
// and per-item time-to-live (TTL) is supported by storing an expiration timestamp
// (Unix epoch seconds) in a configurable field that is enforced on read.
type kvGCPFirestore struct {
	// Project is the GCP project ID that hosts the Firestore database.
	//
	// This is required.
	Project string `json:"project"`
	// Collection is the Firestore collection that documents are read from and
	// written to.
	//
	// This is required.
	Collection string `json:"collection"`
	Attributes struct {
		// Value is the document field where values are read from and written to.
		//
		// This is optional and defaults to "value".
		Value string `json:"value"`
		// TTL is the document field where time-to-live is stored, as a Unix epoch
		// seconds timestamp.
		//
		// This is optional and only required when using SetWithTTL. Expiration is
		// enforced on read: an item whose TTL has passed is reported as a miss.
		TTL string `json:"ttl"`
	} `json:"attributes"`

	// mu guards client, because transforms that share the store can call Setup
	// concurrently.
	mu     sync.Mutex
	client *firestore.Client
}

// newKVGCPFirestore creates a new GCP Firestore KV store.
func newKVGCPFirestore(cfg config.Config) (*kvGCPFirestore, error) {
	var store kvGCPFirestore
	if err := iconfig.Decode(cfg.Settings, &store); err != nil {
		return nil, err
	}

	if store.Project == "" {
		return nil, iconfig.ErrMissingRequiredOption
	}

	return &store, nil
}

func (store *kvGCPFirestore) String() string {
	return toString(store)
}

// isNotFound reports whether err represents a Firestore document-not-found result.
//
// A missing document surfaces as a gRPC codes.NotFound. This is the store's "not
// found" signal: it is a lookup miss, not a failure. Every other error (timeout,
// unavailable, permission denied, and so on) is a real error and must be returned
// to the caller.
func isNotFound(err error) bool {
	return status.Code(err) == codes.NotFound
}

// valueField returns the document field used for values, defaulting to "value"
// when the option is not set.
func (store *kvGCPFirestore) valueField() string {
	if store.Attributes.Value != "" {
		return store.Attributes.Value
	}

	return "value"
}

// doc returns a reference to the document for a key. Firestore document IDs cannot
// contain "/", which would otherwise address a different collection or document.
func (store *kvGCPFirestore) doc(key string) (*firestore.DocumentRef, error) {
	if strings.Contains(key, "/") {
		return nil, fmt.Errorf("kv: gcp_firestore: key %q: %w", key, errInvalidKey)
	}

	return store.client.Collection(store.Collection).Doc(key), nil
}

// Get retrieves a value from the store. A document that does not exist is a miss
// and is returned as (nil, nil). If the item had a time-to-live (TTL) configured
// when it was added and the TTL has passed, then it is also treated as a miss.
func (store *kvGCPFirestore) Get(ctx context.Context, key string) (interface{}, error) {
	ctx = context.WithoutCancel(ctx)

	ref, err := store.doc(key)
	if err != nil {
		return nil, err
	}

	snap, err := ref.Get(ctx)
	if err != nil {
		// A missing document is a lookup miss, not an error.
		if isNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	data := snap.Data()

	if store.expired(data) {
		return nil, nil
	}

	val, found := data[store.valueField()]
	if !found {
		return nil, nil
	}

	return val, nil
}

// Set adds a value to the store, overwriting any existing document for the key.
func (store *kvGCPFirestore) Set(ctx context.Context, key string, val interface{}) error {
	ctx = context.WithoutCancel(ctx)

	data := map[string]interface{}{
		store.valueField(): val,
	}

	ref, err := store.doc(key)
	if err != nil {
		return err
	}

	_, err = ref.Set(ctx, data)

	return err
}

// SetWithTTL adds a value to the store with a time-to-live (TTL), expressed as a
// Unix epoch seconds timestamp. Requires the TTL attribute to be configured.
func (store *kvGCPFirestore) SetWithTTL(ctx context.Context, key string, val interface{}, ttl int64) error {
	if store.Attributes.TTL == "" {
		return iconfig.ErrMissingRequiredOption
	}

	ctx = context.WithoutCancel(ctx)

	data := map[string]interface{}{
		store.valueField():   val,
		store.Attributes.TTL: ttl,
	}

	ref, err := store.doc(key)
	if err != nil {
		return err
	}

	_, err = ref.Set(ctx, data)

	return err
}

// SetAddWithTTL appends a value to a set (unique list) stored in the value field.
// If the document does not exist or has expired, then a new set is created. If the
// TTL attribute is configured, then the TTL field is updated with the provided value.
//
// Uniqueness is provided by Firestore's ArrayUnion, which only adds elements that
// are not already present.
func (store *kvGCPFirestore) SetAddWithTTL(ctx context.Context, key string, val interface{}, ttl int64) error {
	ctx = context.WithoutCancel(ctx)

	// Each element of a list is added as its own member, because Firestore does
	// not allow an array to directly contain another array.
	elems := []interface{}{val}
	if list, ok := val.([]interface{}); ok {
		elems = list
	}

	data := map[string]interface{}{
		store.valueField(): firestore.ArrayUnion(elems...),
	}

	if store.Attributes.TTL != "" {
		data[store.Attributes.TTL] = ttl
	}

	ref, err := store.doc(key)
	if err != nil {
		return err
	}

	// The expiry check and the write run in a transaction, so that members of an
	// expired set are not restored by the new TTL.
	return store.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		snap, err := tx.Get(ref)
		if err != nil && !isNotFound(err) {
			return err
		}

		if err == nil && store.expired(snap.Data()) {
			return tx.Set(ref, data)
		}

		// MergeAll upserts the provided fields, creating the document if it does
		// not already exist while preserving other fields.
		return tx.Set(ref, data, firestore.MergeAll)
	})
}

// expired returns true if the document has a TTL that has passed. A zero value
// indicates that TTL is not configured for the item.
func (store *kvGCPFirestore) expired(data map[string]interface{}) bool {
	if store.Attributes.TTL == "" {
		return false
	}

	ttl := toUnixSeconds(data[store.Attributes.TTL])

	return ttl != 0 && ttl <= time.Now().Unix()
}

// IsEnabled returns true if the Firestore client is ready for use.
func (store *kvGCPFirestore) IsEnabled() bool {
	store.mu.Lock()
	defer store.mu.Unlock()

	return store.client != nil
}

// Setup creates a new Firestore client. When FIRESTORE_EMULATOR_HOST is set, the
// client connects to the emulator automatically.
func (store *kvGCPFirestore) Setup(ctx context.Context) error {
	if store.Project == "" || store.Collection == "" {
		return iconfig.ErrMissingRequiredOption
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	// Avoids unnecessary setup.
	if store.client != nil {
		return nil
	}

	client, err := firestore.NewClient(ctx, store.Project)
	if err != nil {
		return err
	}

	store.client = client

	return nil
}

// Close releases the Firestore client's resources.
func (store *kvGCPFirestore) Close() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if store.client == nil {
		return nil
	}

	err := store.client.Close()
	store.client = nil

	return err
}

// toUnixSeconds coerces a Firestore-decoded TTL field into Unix epoch seconds.
// Firestore returns integers as int64, but values written through other paths may
// decode as other numeric types, so the common cases are handled explicitly.
func toUnixSeconds(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case time.Time:
		return n.Unix()
	default:
		return 0
	}
}
