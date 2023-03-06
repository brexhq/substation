//go:build wasm

package process

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
)

type procKVStore struct {
	process
	Options procKVStoreOptions `json:"options"`
}

type procKVStoreOptions struct{}

func (p procKVStore) String() string {
	return toString(p)
}

func (p procKVStore) Close(ctx context.Context) error {
	return fmt.Errorf("close: kv_store: %v", syscall.ENOSYS)
}

func (p procKVStore) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

func (p procKVStore) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	return capsule, fmt.Errorf("process: kv_store: %v", syscall.ENOSYS)
}
