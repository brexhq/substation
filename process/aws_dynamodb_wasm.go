//go:build wasm

package process

import (
	"context"
	"fmt"
	"syscall"

	"github.com/brexhq/substation/config"
)

type procAWSDynamoDB struct {
	process
	Options procAWSDynamoDBOptions `json:"options"`
}

type procAWSDynamoDBOptions struct{}

func (p procAWSDynamoDB) String() string {
	return toString(p)
}

func (p procAWSDynamoDB) Close(context.Context) error {
	return nil
}

func (p procAWSDynamoDB) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

func (p procAWSDynamoDB) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	return capsule, fmt.Errorf("process: aws_dynamodb: %v", syscall.ENOSYS)
}
