//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

type modKVStore struct{}

func newModKVStore(ctx context.Context, cfg config.Config) (*modKVStore, error) {
	return nil, fmt.Errorf("transform: mod_kv_store: %v", syscall.ENOSYS)
}

func (*modKVStore) String() string {
	return ""
}

func (*modKVStore) Close(context.Context) error {
	return nil
}

func (*modKVStore) Transform(ctx context.Context, messages ...*message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: mod_kv_store: %v", syscall.ENOSYS)
}
