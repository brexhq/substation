//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

type procKVStore struct{}

func newProcKVStore(ctx context.Context, cfg config.Config) (*procKVStore, error) {
	return nil, fmt.Errorf("transform: proc_kv_store: %v", syscall.ENOSYS)
}

func (*procKVStore) String() string {
	return ""
}

func (*procKVStore) Close(context.Context) error {
	return nil
}

func (*procKVStore) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	return nil, fmt.Errorf("transform: proc_kv_store: %v", syscall.ENOSYS)
}
