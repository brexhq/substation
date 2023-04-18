//go:build wasm

package process

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
)

type procHTTP struct {
	process
	Options procHTTPOptions `json:"options"`
}

type procHTTPOptions struct{}

func newProcHTTP(ctx context.Context, cfg config.Config) (p procHTTP, err error) {
	return procHTTP{}, fmt.Errorf("process: http: %v", syscall.ENOSYS)
}

func (p procHTTP) Close(context.Context) error {
	return nil
}

func (p procHTTP) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

func (p procHTTP) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	return capsule, fmt.Errorf("process: http: %v", syscall.ENOSYS)
}
