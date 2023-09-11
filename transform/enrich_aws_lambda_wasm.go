//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newEnrichAWSLambda(context.Context, config.Config) (*modAWSLambda, error) {
	return nil, fmt.Errorf("process: enrich_aws_lambda: %v", syscall.ENOSYS)
}

type enrichAWSLambda struct{}

func (*enrichAWSLambda) String() string {
	return ""
}

func (*enrichAWSLambda) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: enrich_aws_lambda: %v", syscall.ENOSYS)
}
