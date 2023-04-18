//go:build wasm

package process

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
)

type procAWSLambda struct {
	process
	Options procAWSLambdaOptions `json:"options"`
}

type procAWSLambdaOptions struct{}

func newProcAWSLambda(ctx context.Context, cfg config.Config) (p procAWSLambda, err error) {
	return procAWSLambda{}, fmt.Errorf("process: aws_lambda: %v", syscall.ENOSYS)
}

func (p procAWSLambda) String() string {
	return toString(p)
}

func (p procAWSLambda) Close(context.Context) error {
	return nil
}

func (p procAWSLambda) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

func (p procAWSLambda) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	return capsule, fmt.Errorf("process: aws_lambda: %v", syscall.ENOSYS)
}
