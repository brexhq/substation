//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

type procAWSDynamoDB struct{}

func newProcAWSDynamoDB(context.Context, config.Config) (*procAWSDynamoDB, error) {
	return nil, fmt.Errorf("transform: proc_aws_dynamodb: %v", syscall.ENOSYS)
}

func (*procAWSDynamoDB) String() string {
	return ""
}

func (*procAWSDynamoDB) Close(context.Context) error {
	return nil
}

func (*procAWSDynamoDB) Transform(context.Context, ...*mess.Message) ([]*mess.Message, error) {
	return nil, fmt.Errorf("transform: proc_aws_dynamodb: %v", syscall.ENOSYS)
}
