//go:build wasm

package process

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
)

type procIPDatabase struct {
	process
	Options config.Config `json:"options"`
}

func newProcIPDatabase(cfg config.Config) (p procIPDatabase, err error) {
	return procIPDatabase{}, fmt.Errorf("process: ip_database: %v", syscall.ENOSYS)
}

func (p procIPDatabase) String() string {
	return toString(p)
}

func (p procIPDatabase) Close(ctx context.Context) error {
	return fmt.Errorf("close: ip_database: %v", syscall.ENOSYS)
}

func (p procIPDatabase) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

func (p procIPDatabase) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	return capsule, fmt.Errorf("process: ip_database: %v", syscall.ENOSYS)
}
