//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

type modDNS struct{}

func newModDNS(ctx context.Context, cfg config.Config) (*modDNS, error) {
	return nil, fmt.Errorf("transform: proc_dns: %v", syscall.ENOSYS)
}

func (*modDNS) String() string {
	return ""
}

func (*modDNS) Close(context.Context) error {
	return nil
}

func (*modDNS) Transform(_ context.Context, _ ...*message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: proc_dns: %v", syscall.ENOSYS)
}
