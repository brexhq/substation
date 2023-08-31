//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

type modAWSDynamoDB struct{}

func newModAWSDynamoDB(context.Context, config.Config) (*modAWSDynamoDB, error) {
	return nil, fmt.Errorf("transform: mod_aws_dynamodb: %v", syscall.ENOSYS)
}

func (*modAWSDynamoDB) String() string {
	return ""
}

func (*modAWSDynamoDB) Close(context.Context) error {
	return nil
}

func (*modAWSDynamoDB) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: mod_aws_dynamodb: %v", syscall.ENOSYS)
}
