//go:build wasm

package transform

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newEnrichAWSDynamoDB(context.Context, config.Config) (*modAWSDynamoDB, error) {
	return nil, fmt.Errorf("transform: new_enrich_aws_dynamodb: %v", syscall.ENOSYS)
}

type enrichAWSDynamoDB struct{}

func (*enrichAWSDynamoDB) String() string {
	return ""
}

func (*enrichAWSDynamoDB) Close(context.Context) error {
	return nil
}

func (*enrichAWSDynamoDB) Transform(context.Context, *message.Message) ([]*message.Message, error) {
	return nil, fmt.Errorf("transform: enrich_aws_dynamodb: %v", syscall.ENOSYS)
}
