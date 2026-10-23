package kv

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

// TestGCPFirestoreConfigDecode verifies that a config is decoded into the store's
// fields using the standard iconfig.Decode idiom.
func TestGCPFirestoreConfigDecode(t *testing.T) {
	cfg := config.Config{
		Type: "gcp_firestore",
		Settings: map[string]interface{}{
			"project":    "my-project",
			"collection": "my-collection",
			"attributes": map[string]interface{}{
				"value": "value",
				"ttl":   "ttl",
			},
		},
	}

	store, err := newKVGCPFirestore(cfg)
	if err != nil {
		t.Fatalf("newKVGCPFirestore: unexpected error: %v", err)
	}

	if store.Project != "my-project" {
		t.Errorf("Project = %q, want %q", store.Project, "my-project")
	}
	if store.Collection != "my-collection" {
		t.Errorf("Collection = %q, want %q", store.Collection, "my-collection")
	}
	if store.Attributes.Value != "value" {
		t.Errorf("Attributes.Value = %q, want %q", store.Attributes.Value, "value")
	}
	if store.Attributes.TTL != "ttl" {
		t.Errorf("Attributes.TTL = %q, want %q", store.Attributes.TTL, "ttl")
	}
}

// TestGCPFirestoreConfigValidate verifies that the constructor rejects a config
// that is missing the required project.
func TestGCPFirestoreConfigValidate(t *testing.T) {
	cfg := config.Config{
		Type: "gcp_firestore",
		Settings: map[string]interface{}{
			"collection": "my-collection",
		},
	}

	if _, err := newKVGCPFirestore(cfg); !errors.Is(err, iconfig.ErrMissingRequiredOption) {
		t.Errorf("newKVGCPFirestore: err = %v, want %v", err, iconfig.ErrMissingRequiredOption)
	}
}

// TestGCPFirestoreSetupValidate verifies that Setup rejects a config that is
// missing the required collection (project is validated by the constructor).
func TestGCPFirestoreSetupValidate(t *testing.T) {
	store := &kvGCPFirestore{Project: "my-project"}

	if err := store.Setup(context.Background()); !errors.Is(err, iconfig.ErrMissingRequiredOption) {
		t.Errorf("Setup: err = %v, want %v", err, iconfig.ErrMissingRequiredOption)
	}
}

// TestGCPFirestoreSetWithTTLRequiresField verifies that SetWithTTL fails when the
// TTL attribute is not configured, mirroring the DynamoDB backend.
func TestGCPFirestoreSetWithTTLRequiresField(t *testing.T) {
	store := &kvGCPFirestore{Project: "my-project", Collection: "my-collection"}

	if err := store.SetWithTTL(context.Background(), "key", "val", 1); !errors.Is(err, iconfig.ErrMissingRequiredOption) {
		t.Errorf("SetWithTTL: err = %v, want %v", err, iconfig.ErrMissingRequiredOption)
	}
}

// TestGCPFirestoreInvalidKey verifies that keys containing "/" are rejected
// before any request is made.
func TestGCPFirestoreInvalidKey(t *testing.T) {
	store := &kvGCPFirestore{Project: "my-project", Collection: "my-collection"}
	ctx := context.Background()

	if _, err := store.Get(ctx, "a/b"); !errors.Is(err, errInvalidKey) {
		t.Errorf("Get: err = %v, want %v", err, errInvalidKey)
	}
	if err := store.Set(ctx, "a/b", "val"); !errors.Is(err, errInvalidKey) {
		t.Errorf("Set: err = %v, want %v", err, errInvalidKey)
	}
	if err := store.SetAddWithTTL(ctx, "a/b", "val", 1); !errors.Is(err, errInvalidKey) {
		t.Errorf("SetAddWithTTL: err = %v, want %v", err, errInvalidKey)
	}
}

// TestGCPFirestoreIsNotFound is the core of the miss-vs-error contract: a Firestore
// document-not-found surfaces as a gRPC codes.NotFound and must be translated into a
// lookup miss, while every other failure must remain an error.
func TestGCPFirestoreIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"not found is a miss", status.Error(codes.NotFound, "document not found"), true},
		{"unavailable is an error", status.Error(codes.Unavailable, "backend unavailable"), false},
		{"permission denied is an error", status.Error(codes.PermissionDenied, "denied"), false},
		{"deadline exceeded is an error", status.Error(codes.DeadlineExceeded, "timeout"), false},
		{"plain error is an error", errors.New("boom"), false},
		{"nil is not a miss", nil, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isNotFound(tc.err); got != tc.want {
				t.Errorf("isNotFound(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// newEmulatorStore returns a store connected to the Firestore emulator. It skips
// the test when FIRESTORE_EMULATOR_HOST is unset so that the unit test suite
// passes without the emulator.
func newEmulatorStore(t *testing.T) *kvGCPFirestore {
	t.Helper()

	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST is not set; skipping Firestore emulator integration test")
	}

	cfg := config.Config{
		Type: "gcp_firestore",
		Settings: map[string]interface{}{
			"project": "substation-test",
			// A collection per run keeps data left in the emulator from
			// affecting the results.
			"collection": fmt.Sprintf("kv_test_%d", time.Now().UnixNano()),
			"attributes": map[string]interface{}{
				"value": "value",
				"ttl":   "ttl",
			},
		},
	}

	store, err := newKVGCPFirestore(cfg)
	if err != nil {
		t.Fatalf("newKVGCPFirestore: %v", err)
	}

	if err := store.Setup(context.Background()); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if !store.IsEnabled() {
		t.Fatal("IsEnabled: got false, want true after Setup")
	}

	return store
}

// TestGCPFirestoreEmulator exercises the get and set round-trip, including TTL,
// against the Firestore emulator.
func TestGCPFirestoreEmulator(t *testing.T) {
	store := newEmulatorStore(t)
	ctx := context.Background()

	// A missing key is a miss, not an error.
	got, err := store.Get(ctx, "missing")
	if err != nil {
		t.Fatalf("Get(missing): unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("Get(missing) = %v, want nil (miss)", got)
	}

	// put -> get round-trip.
	if err := store.Set(ctx, "foo", "bar"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err = store.Get(ctx, "foo")
	if err != nil {
		t.Fatalf("Get(foo): %v", err)
	}
	if got != "bar" {
		t.Errorf("Get(foo) = %v, want %q", got, "bar")
	}

	// TTL honored: an expired value reads back as a miss.
	expired := time.Now().Add(-1 * time.Hour).Unix()
	if err := store.SetWithTTL(ctx, "stale", "gone", expired); err != nil {
		t.Fatalf("SetWithTTL(expired): %v", err)
	}

	got, err = store.Get(ctx, "stale")
	if err != nil {
		t.Fatalf("Get(stale): %v", err)
	}
	if got != nil {
		t.Errorf("Get(stale) = %v, want nil (expired miss)", got)
	}

	// A future TTL is still readable.
	future := time.Now().Add(1 * time.Hour).Unix()
	if err := store.SetWithTTL(ctx, "fresh", "here", future); err != nil {
		t.Fatalf("SetWithTTL(future): %v", err)
	}

	got, err = store.Get(ctx, "fresh")
	if err != nil {
		t.Fatalf("Get(fresh): %v", err)
	}
	if got != "here" {
		t.Errorf("Get(fresh) = %v, want %q", got, "here")
	}
}

// TestGCPFirestoreEmulatorSetAdd exercises SetAddWithTTL against the Firestore
// emulator.
func TestGCPFirestoreEmulatorSetAdd(t *testing.T) {
	store := newEmulatorStore(t)
	ctx := context.Background()
	expired := time.Now().Add(-1 * time.Hour).Unix()
	future := time.Now().Add(1 * time.Hour).Unix()

	// SetAddWithTTL builds a unique set.
	if err := store.SetAddWithTTL(ctx, "set", "a", future); err != nil {
		t.Fatalf("SetAddWithTTL(a): %v", err)
	}
	if err := store.SetAddWithTTL(ctx, "set", "a", future); err != nil {
		t.Fatalf("SetAddWithTTL(a again): %v", err)
	}
	if err := store.SetAddWithTTL(ctx, "set", "b", future); err != nil {
		t.Fatalf("SetAddWithTTL(b): %v", err)
	}

	got, err := store.Get(ctx, "set")
	if err != nil {
		t.Fatalf("Get(set): %v", err)
	}
	members, ok := got.([]interface{})
	if !ok {
		t.Fatalf("Get(set) = %T, want []interface{}", got)
	}
	if len(members) != 2 {
		t.Errorf("Get(set) has %d members, want 2 (unique set): %v", len(members), members)
	}

	// Members of an expired set are not restored by a later add.
	if err := store.SetAddWithTTL(ctx, "expired_set", "a", expired); err != nil {
		t.Fatalf("SetAddWithTTL(expired a): %v", err)
	}
	if err := store.SetAddWithTTL(ctx, "expired_set", "b", future); err != nil {
		t.Fatalf("SetAddWithTTL(b): %v", err)
	}

	got, err = store.Get(ctx, "expired_set")
	if err != nil {
		t.Fatalf("Get(expired_set): %v", err)
	}
	if members, ok := got.([]interface{}); !ok || len(members) != 1 || members[0] != "b" {
		t.Errorf("Get(expired_set) = %v, want [b]", got)
	}

	// A list is added as individual members.
	if err := store.SetAddWithTTL(ctx, "list", []interface{}{"x", "y"}, future); err != nil {
		t.Fatalf("SetAddWithTTL(list): %v", err)
	}
	if err := store.SetAddWithTTL(ctx, "list", []interface{}{"y", "z"}, future); err != nil {
		t.Fatalf("SetAddWithTTL(list again): %v", err)
	}

	got, err = store.Get(ctx, "list")
	if err != nil {
		t.Fatalf("Get(list): %v", err)
	}
	if members, ok := got.([]interface{}); !ok || len(members) != 3 {
		t.Errorf("Get(list) = %v, want 3 unique members", got)
	}
}
