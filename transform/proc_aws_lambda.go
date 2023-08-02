//go:build !wasm

package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/lambda"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
)

// errProcAWSLambdaInputNotAnObject is returned when the input is not an object.
var errProcAWSLambdaInputNotAnObject = fmt.Errorf("input is not an object")

type procAWSLambdaConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// IgnoreErrors indicates if errors returned by the AWS Lambda function should be ignored.
	//
	// This is optional and defaults to false.
	// TODO(v1.0): Change to ErrorOnFailure.
	IgnoreErrors bool `json:"ignore_errors"`
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
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Key == "" || conf.SetKey == "" {
		return nil, fmt.Errorf("transform: proc_aws_lambda: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.FunctionName == "" {
		return nil, fmt.Errorf("transform: proc_aws_lambda: options %+v: %v", conf, errors.ErrMissingRequiredOption)
	}

	proc := procAWSLambda{
		conf: conf,
	}

	// Setup the AWS client.
	if !proc.client.IsEnabled() {
		proc.client.Setup()
	}

	return &proc, nil
}

func (t *procAWSLambda) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procAWSLambda) Close(context.Context) error {
	return nil
}

func (t *procAWSLambda) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		result := message.Get(t.conf.Key)
		if !result.IsObject() {
			return nil, fmt.Errorf("transform: proc_aws_lambda: key %s: %v", t.conf.Key, errProcAWSLambdaInputNotAnObject)
		}

		resp, err := t.client.Invoke(ctx, t.conf.FunctionName, []byte(result.Raw))
		if err != nil {
			return nil, fmt.Errorf("transform: proc_aws_lambda: %v", err)
		}

		if resp.FunctionError != nil && !t.conf.IgnoreErrors {
			resErr := json.Get(resp.Payload, "errorMessage").String()
			return nil, fmt.Errorf("transform: proc_aws_lambda: %v", resErr)
		}

		if resp.FunctionError != nil {
			output = append(output, message)
			continue
		}

		if err := message.Set(t.conf.SetKey, resp.Payload); err != nil {
			return nil, fmt.Errorf("transform: proc_aws_lambda: %v", err)
		}

		output = append(output, message)
	}

	return output, nil
}
