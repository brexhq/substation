package kv_test

import (
	"context"
	"fmt"
	"time"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/internal/kv"
)

func Example_memory() {
	ctx := context.TODO()

	// create KV config
	cfg := config.Config{
		Type: "memory",
		Settings: map[string]interface{}{
			"capacity": 3,
		},
	}

	// get KV store using factory method
	kvStore, err := kv.Get(cfg)
	if err != nil {
		panic(err)
	}

	// setup and defer closing KV store
	if err := kvStore.Setup(ctx); err != nil {
		panic(err)
	}
	defer kvStore.Close()

	// set a series of values in the store
	if err := kvStore.Set(ctx, "foo", "bar"); err != nil {
		panic(err)
	}

	if err := kvStore.Set(ctx, "baz", "qux"); err != nil {
		panic(err)
	}

	if err := kvStore.Set(ctx, "quux", "corge"); err != nil {
		panic(err)
	}

	// retrieve a value from the store
	item, err := kvStore.Get(ctx, "foo")
	if err != nil {
		panic(err)
	}

	// Output: bar
	fmt.Println(item)
}

func Example_memoryWithTTL() {
	ctx := context.TODO()

	// create KV config
	cfg := config.Config{
		Type: "memory",
		Settings: map[string]interface{}{
			"capacity": 1,
		},
	}

	// get KV store using factory method
	kvStore, err := kv.Get(cfg)
	if err != nil {
		panic(err)
	}

	// setup and defer closing KV store
	if err := kvStore.Setup(ctx); err != nil {
		panic(err)
	}
	defer kvStore.Close()

	// set a value with time-to-live enabled in the KV store
	ttl := time.Now().Add(1 * time.Microsecond).Unix()
	if err := kvStore.SetWithTTL(ctx, "foo", "bar", ttl); err != nil {
		panic(err)
	}

	time.Sleep(15 * time.Microsecond)

	// retrieve a value from the store
	// if the time-to-live has passed, then the item has expired and is
	// no longer available
	item, err := kvStore.Get(ctx, "foo")
	if err != nil {
		panic(err)
	}

	// Output: <nil>
	fmt.Println(item)
}
