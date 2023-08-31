//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

type modHTTP struct{}

func newModHTTP(ctx context.Context, cfg config.Config) (*modHTTP, error) {
	return nil, fmt.Errorf("transform: proc_http: %v", syscall.ENOSYS)
}

func (*modHTTP) String() string {
	return ""
}

func (*modHTTP) Close(context.Context) error {
	return nil
}

func (*modHTTP) Transform(ctx context.Context, messages ...*message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: proc_http: %v", syscall.ENOSYS)
}
