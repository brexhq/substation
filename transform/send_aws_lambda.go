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
)

// Payloads greater than 256 KB in size cannot be
// sent to an AWS Lambda function.
const sendLambdaPayloadSizeLimit = 1024 * 1024 * 256

// errSendLambdaPayloadSizeLimit is returned when data exceeds the Lambda
// payload size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendLambdaPayloadSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSLambdaConfig struct {
	AWS   iconfig.AWS   `json:"aws"`
	Retry iconfig.Retry `json:"retry"`

	// FunctionName is the AWS Lambda function to asynchronously invoke.
	FunctionName string `json:"function_name"`
}

func (c *sendAWSLambdaConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSLambdaConfig) Validate() error {
	if c.FunctionName == "" {
		return fmt.Errorf("function_name: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSLambda(_ context.Context, cfg config.Config) (*sendAWSLambda, error) {
	conf := sendAWSLambdaConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: send_aws_lambda: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: send_aws_lambda: %v", err)
	}

	tf := sendAWSLambda{
		conf:     conf,
		function: conf.FunctionName,
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:        conf.AWS.Region,
		AssumeRoleARN: conf.AWS.AssumeRoleARN,
		MaxRetries:    conf.Retry.Count,
	})

	return &tf, nil
}

type sendAWSLambda struct {
	conf     sendAWSLambdaConfig
	function string

	// client is safe for concurrent use.
	client lambda.API
}

func (tf *sendAWSLambda) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendLambdaPayloadSizeLimit {
		return nil, fmt.Errorf("transform: send_aws_lambda: %v", errSendLambdaPayloadSizeLimit)
	}

	// Invoke the AWS Lambda function.
	if _, err := tf.client.InvokeAsync(ctx, tf.function, msg.Data()); err != nil {
		return nil, fmt.Errorf("transform: send_aws_lambda: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *sendAWSLambda) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
