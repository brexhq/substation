//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

type procHTTP struct{}

func newProcHTTP(ctx context.Context, cfg config.Config) (*procHTTP, error) {
	return nil, fmt.Errorf("transform: proc_http: %v", syscall.ENOSYS)
}

func (*procHTTP) String() string {
	return ""
}

func (*procHTTP) Close(context.Context) error {
	return nil
}

func (*procHTTP) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	return nil, fmt.Errorf("transform: proc_http: %v", syscall.ENOSYS)
}
