//go:build wasm

package process

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
)

type procDNS struct {
	process
	Options procDNSOptions `json:"options"`
}

type procDNSOptions struct{}

func (p procDNS) Close(context.Context) error {
	return nil
}

func (p procDNS) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

func (p procDNS) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	return capsule, fmt.Errorf("process: dns: %v", syscall.ENOSYS)
}
