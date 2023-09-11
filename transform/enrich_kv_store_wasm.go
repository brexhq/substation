//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newEnrichKVStoreGet(context.Context, config.Config) (*enrichKVStoreGet, error) {
	return nil, fmt.Errorf("transform: enrich_kv_store: %v", syscall.ENOSYS)
}

type enrichKVStoreGet struct{}

func (*enrichKVStoreGet) String() string {
	return ""
}

func (*enrichKVStoreGet) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: enrich_kv_store: %v", syscall.ENOSYS)
}

func newEnrichKVStoreSet(context.Context, config.Config) (*enrichKVStoreSet, error) {
	return nil, fmt.Errorf("transform: enrich_kv_store: %v", syscall.ENOSYS)
}

type enrichKVStoreSet struct{}

func (*enrichKVStoreSet) String() string {
	return ""
}

func (*enrichKVStoreSet) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: enrich_kv_store: %v", syscall.ENOSYS)
}
