//go:build !wasm

package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/lambda"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

var lambdaAPI lambda.API

// errAWSLambdaInputNotAnObject is returned when the input is not a JSON object.
const errAWSLambdaInputNotAnObject = errors.Error("input is not an object")

// awsLambda processes data by synchronously invoking an AWS Lambda function
// and returning the payload. The average latency of synchronously invoking
// a function is 10s of milliseconds, but latency can take 100s to 1000s of
// milliseconds depending on the function and may have significant impact on
// end-to-end data processing latency. If Substation is running in AWS Lambda
// with Kinesis, then this latency can be mitigated by increasing the parallelization
// factor of the Lambda
// (https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html).
//
// This processor supports the object handling pattern.
type procAWSLambda struct {
	process
	Options procAWSLambdaOptions `json:"options"`
}

type procAWSLambdaOptions struct {
	// FunctionName is the AWS Lambda function to synchronously invoke.
	FunctionName string `json:"function_name"`
}

// String returns the processor settings as an object.
func (p procAWSLambda) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procAWSLambda) Close(context.Context) error {
	return nil
}

// Create a new AWS Lambda processor.
func newProcAWSLambda(ctx context.Context, cfg config.Config) (p procAWSLambda, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procAWSLambda{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procAWSLambda{}, err
	}

	// fail if required options are missing
	if p.Options.FunctionName == "" {
		return procAWSLambda{}, fmt.Errorf("process: aws_lambda: options %+v: %v", p.Options, errors.ErrMissingRequiredOption)
	}

	return p, nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procAWSLambda) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procAWSLambda) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// lazy load API
	if !lambdaAPI.IsEnabled() {
		lambdaAPI.Setup()
	}

	result := capsule.Get(p.Key)
	if !result.IsObject() {
		return capsule, fmt.Errorf("process: aws_lambda: key %s: %v", p.Key, errAWSLambdaInputNotAnObject)
	}

	resp, err := lambdaAPI.Invoke(ctx, p.Options.FunctionName, []byte(result.Raw))
	if err != nil {
		return capsule, fmt.Errorf("process: aws_lambda: %v", err)
	}

	if resp.FunctionError != nil && !p.IgnoreErrors {
		resErr := json.Get(resp.Payload, "errorMessage").String()
		return capsule, fmt.Errorf("process: aws_lambda: %v", resErr)
	}

	if resp.FunctionError != nil {
		return capsule, nil
	}

	if err := capsule.Set(p.SetKey, resp.Payload); err != nil {
		return capsule, fmt.Errorf("process: aws_lambda: %v", err)
	}

	return capsule, nil
}
