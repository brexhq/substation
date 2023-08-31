//go:build !wasm

package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/lambda"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/message"
)

// errModAWSLambdaInputNotAnObject is returned when the input is not an object.
var errModAWSLambdaInputNotAnObject = fmt.Errorf("input is not an object")

type modAWSLambdaConfig struct {
	Object configObject `json:"object"`
	AWS    configAWS    `json:"aws"`
	Retry  configRetry  `json:"retry"`

	// ErrorOnFailure determines whether an error is returned during processing.
	//
	// This is optional and defaults to false.
	ErrorOnFailure bool `json:"error_on_failure"`
	// FunctionName is the AWS Lambda function to synchronously invoke.
	FunctionName string `json:"function_name"`
}

type modAWSLambda struct {
	conf modAWSLambdaConfig

	// client is safe for concurrent access.
	client lambda.API
}

func newModAWSLambda(ctx context.Context, cfg config.Config) (*modAWSLambda, error) {
	conf := modAWSLambdaConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_aws_lambda: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" {
		return nil, fmt.Errorf("transform: new_mod_aws_lambda: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_aws_lambda: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.FunctionName == "" {
		return nil, fmt.Errorf("transform: new_mod_aws_lambda: function_name: %v", errors.ErrMissingRequiredOption)
	}

	tf := modAWSLambda{
		conf: conf,
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		AssumeRole: conf.AWS.AssumeRole,
		MaxRetries: conf.Retry.Attempts,
	})

	return &tf, nil
}

func (tf *modAWSLambda) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modAWSLambda) Close(context.Context) error {
	return nil
}

func (tf *modAWSLambda) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	result := msg.GetObject(tf.conf.Object.Key)
	if !result.IsObject() {
		return nil, fmt.Errorf("transform: mod_aws_lambda: key %s: %v", tf.conf.Object.Key, errModAWSLambdaInputNotAnObject)
	}

	resp, err := tf.client.Invoke(ctx, tf.conf.FunctionName, result.RawBytes())
	if err != nil {
		return nil, fmt.Errorf("transform: mod_aws_lambda: %v", err)
	}

	// If ErrorOnFailure is configured, then errors are returned,
	// but otherwise the message is returned as-is.
	if resp.FunctionError != nil && tf.conf.ErrorOnFailure {
		resErr := json.Get(resp.Payload, "errorMessage").String()
		return nil, fmt.Errorf("transform: mod_aws_lambda: %v", resErr)
	} else if resp.FunctionError != nil {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, resp.Payload); err != nil {
		return nil, fmt.Errorf("transform: mod_aws_lambda: %v", err)
	}

	return []*message.Message{msg}, nil
}
