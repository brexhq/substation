//go:build wasm

package kv

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
)

type kvMMDB struct{}

func newKVMMDB(config.Config) (*kvMMDB, error) {
	return nil, fmt.Errorf("kv: mmdb: %v", syscall.ENOSYS)
}

func (*kvMMDB) String() string {
	return ""
}

func (*kvMMDB) Get(context.Context, string) (interface{}, error) {
	return nil, fmt.Errorf("kv: mmdb: %v", syscall.ENOSYS)
}

func (*kvMMDB) Set(context.Context, string, interface{}) error {
	return fmt.Errorf("kv: mmdb: %v", syscall.ENOSYS)
}

func (*kvMMDB) SetWithTTL(context.Context, string, interface{}, int64) error {
	return fmt.Errorf("kv: mmdb: %v", syscall.ENOSYS)
}

func (*kvMMDB) IsEnabled() bool {
	return false
}

func (*kvMMDB) Setup(context.Context) error {
	return fmt.Errorf("kv: mmdb: %v", syscall.ENOSYS)
}

func (*kvMMDB) Close() error {
	return fmt.Errorf("kv: mmdb: %v", syscall.ENOSYS)
}
