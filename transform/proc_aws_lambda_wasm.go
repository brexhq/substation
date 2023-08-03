//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

type procAWSLambda struct{}

func newProcAWSLambda(context.Context, config.Config) (*procAWSLambda, error) {
	return nil, fmt.Errorf("process: aws_lambda: %v", syscall.ENOSYS)
}

func (*procAWSLambda) String() string {
	return ""
}

func (*procAWSLambda) Close(context.Context) error {
	return nil
}

func (*procAWSLambda) Transform(context.Context, ...*mess.Message) ([]*mess.Message, error) {
	return nil, fmt.Errorf("transform: proc_aws_lambda: %v", syscall.ENOSYS)
}
