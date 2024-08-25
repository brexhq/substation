package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/tidwall/gjson"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iaws "github.com/brexhq/substation/v2/internal/aws"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	ierrors "github.com/brexhq/substation/v2/internal/errors"
)

type enrichAWSLambdaConfig struct {
	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	AWS    iconfig.AWS    `json:"aws"`
}

func (c *enrichAWSLambdaConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *enrichAWSLambdaConfig) Validate() error {
	if c.Object.SourceKey == "" {
		return fmt.Errorf("object_source_key: %v", ierrors.ErrMissingRequiredOption)
	}

	if c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", ierrors.ErrMissingRequiredOption)
	}

	if c.AWS.ARN == "" {
		return fmt.Errorf("aws.arn: %v", ierrors.ErrMissingRequiredOption)
	}

	return nil
}

func newEnrichAWSLambda(ctx context.Context, cfg config.Config) (*enrichAWSLambda, error) {
	conf := enrichAWSLambdaConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform enrich_aws_lambda: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "enrich_aws_lambda"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := enrichAWSLambda{
		conf: conf,
	}

	// Setup the AWS client.
	awsCfg, err := iaws.New(ctx, iaws.Config{
		Region:  iaws.ParseRegion(conf.AWS.ARN),
		RoleARN: conf.AWS.AssumeRoleARN,
	})
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf.client = lambda.NewFromConfig(awsCfg)

	return &tf, nil
}

type enrichAWSLambda struct {
	conf   enrichAWSLambdaConfig
	client *lambda.Client
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
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errMsgInvalidObject)
	}

	input := &lambda.InvokeInput{
		FunctionName: &tf.conf.AWS.ARN,
		Payload:      value.Bytes(),
	}

	resp, err := tf.client.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	if resp.FunctionError != nil {
		resErr := gjson.GetBytes(resp.Payload, "errorMessage").String()
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, resErr)
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, resp.Payload); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *enrichAWSLambda) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
