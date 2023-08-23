//go:build !wasm

package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/lambda"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
)

// errprocAWSLambdaInputNotAnObject is returned when the input is not an object.
var errprocAWSLambdaInputNotAnObject = fmt.Errorf("input is not an object")

type procAWSLambdaConfig struct {
	Auth    _config.ConfigAWSAuth `json:"auth"`
	Request _config.ConfigRequest `json:"request"`
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// ErrorOnFailure determines whether an error is returned during processing.
	//
	// This is optional and defaults to false.
	ErrorOnFailure bool `json:"error_on_failure"`
	// FunctionName is the AWS Lambda function to synchronously invoke.
	FunctionName string `json:"function_name"`
}

type procAWSLambda struct {
	conf procAWSLambdaConfig

	// client is safe for concurrent access.
	client lambda.API
}

func newProcAWSLambda(ctx context.Context, cfg config.Config) (*procAWSLambda, error) {
	conf := procAWSLambdaConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Key == "" || conf.SetKey == "" {
		return nil, fmt.Errorf("transform: proc_aws_lambda: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.FunctionName == "" {
		return nil, fmt.Errorf("transform: proc_aws_lambda: options %+v: %v", conf, errors.ErrMissingRequiredOption)
	}

	mod := procAWSLambda{
		conf: conf,
	}

	// Setup the AWS client.
	mod.client.Setup(aws.Config{
		Region:     conf.Auth.Region,
		AssumeRole: conf.Auth.AssumeRole,
		MaxRetries: conf.Request.MaxRetries,
	})

	return &mod, nil
}

func (mod *procAWSLambda) String() string {
	b, _ := gojson.Marshal(mod.conf)
	return string(b)
}

func (*procAWSLambda) Close(context.Context) error {
	return nil
}

func (mod *procAWSLambda) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	result := message.Get(mod.conf.Key)
	if !result.IsObject() {
		return nil, fmt.Errorf("transform: proc_aws_lambda: key %s: %v", mod.conf.Key, errprocAWSLambdaInputNotAnObject)
	}

	resp, err := mod.client.Invoke(ctx, mod.conf.FunctionName, []byte(result.Raw))
	if err != nil {
		return nil, fmt.Errorf("transform: proc_aws_lambda: %v", err)
	}

	// If ErrorOnFailure is configured, then errors are returned,
	// but otherwise the message is returned as-is.
	if resp.FunctionError != nil && mod.conf.ErrorOnFailure {
		resErr := json.Get(resp.Payload, "errorMessage").String()
		return nil, fmt.Errorf("transform: proc_aws_lambda: %v", resErr)
	} else if resp.FunctionError != nil {
		return []*mess.Message{message}, nil
	}

	if err := message.Set(mod.conf.SetKey, resp.Payload); err != nil {
		return nil, fmt.Errorf("transform: proc_aws_lambda: %v", err)
	}

	return []*mess.Message{message}, nil
}
