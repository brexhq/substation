//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

type modAWSLambda struct{}

func newModAWSLambda(context.Context, config.Config) (*modAWSLambda, error) {
	return nil, fmt.Errorf("process: mod_aws_lambda: %v", syscall.ENOSYS)
}

func (*modAWSLambda) String() string {
	return ""
}

func (*modAWSLambda) Close(context.Context) error {
	return nil
}

func (*modAWSLambda) Transform(context.Context, ...*message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: mod_aws_lambda: %v", syscall.ENOSYS)
}
