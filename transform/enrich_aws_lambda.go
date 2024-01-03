package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/lambda"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
	"github.com/tidwall/gjson"
)

type enrichAWSLambdaConfig struct {
	Object iconfig.Object `json:"object"`
	AWS    iconfig.AWS    `json:"aws"`
	Retry  iconfig.Retry  `json:"retry"`

	// FunctionName is the AWS Lambda function to synchronously invoke.
	FunctionName string `json:"function_name"`
}

func (c *enrichAWSLambdaConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *enrichAWSLambdaConfig) Validate() error {
	if c.Object.SourceKey == "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.FunctionName == "" {
		return fmt.Errorf("function_name: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newEnrichAWSLambda(_ context.Context, cfg config.Config) (*enrichAWSLambda, error) {
	conf := enrichAWSLambdaConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: enrich_aws_lambda: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: enrich_aws_lambda: %v", err)
	}

	tf := enrichAWSLambda{
		conf: conf,
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		RoleARN:    conf.AWS.RoleARN,
		MaxRetries: conf.Retry.Count,
	})

	return &tf, nil
}

type enrichAWSLambda struct {
	conf enrichAWSLambdaConfig

	// client is safe for concurrent access.
	client lambda.API
}

func (tf *enrichAWSLambda) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if !json.Valid(value.Bytes()) {
		return nil, fmt.Errorf("transform: enrich_aws_lambda: %v", errMsgInvalidObject)
	}

	resp, err := tf.client.Invoke(ctx, tf.conf.FunctionName, value.Bytes())
	if err != nil {
		return nil, fmt.Errorf("transform: enrich_aws_lambda: %v", err)
	}

	if resp.FunctionError != nil {
		resErr := gjson.GetBytes(resp.Payload, "errorMessage").String()
		return nil, fmt.Errorf("transform: enrich_aws_lambda: %v", resErr)
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, resp.Payload); err != nil {
		return nil, fmt.Errorf("transform: enrich_aws_lambda: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *enrichAWSLambda) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
