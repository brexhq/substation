//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

type procDNS struct{}

func newProcDNS(ctx context.Context, cfg config.Config) (*procDNS, error) {
	return nil, fmt.Errorf("transform: proc_dns: %v", syscall.ENOSYS)
}

func (*procDNS) String() string {
	return ""
}

func (*procDNS) Close(context.Context) error {
	return nil
}

func (*procDNS) Transform(_ context.Context, _ ...*mess.Message) ([]*mess.Message, error) {
	return nil, fmt.Errorf("transform: proc_dns: %v", syscall.ENOSYS)
}
