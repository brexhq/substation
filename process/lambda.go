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

// errLambdaInputNotAnObject is returned when the input is not a JSON object.
const errLambdaInputNotAnObject = errors.Error("input is not an object")

/*
Lambda processes data by synchronously invoking an AWS Lambda and returning the payload. The average latency of synchronously invoking a Lambda function is 10s of milliseconds, but latency can take 100s to 1000s of milliseconds depending on the function which can have significant impact on total event latency. If Substation is running in AWS Lambda with Kinesis, then this latency can be mitigated by increasing the parallelization factor of the Lambda (https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html).

The input key's value must be a JSON object that contains settings for the Lambda. It is recommended to use the copy and insert processors to create the JSON object before calling this processor and to use the delete processor to remove the JSON object after calling this processor.

The processor supports these patterns:
	JSON:
		{"foo":"bar","lambda":{"lookup":"baz"}} >>> {"foo":"bar","lambda":{"baz":"qux"}}

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "lambda",
		"settings": {
			"options": {
				"function": "foo-function"
			},
			"input_key": "lambda",
			"output_key": "lambda"
		}
	}
*/
type Lambda struct {
	Options   LambdaOptions    `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
LambdaOptions contains custom options settings for the Lambda processor:
	Function:
		Lambda function to invoke
	ErrorOnFailure (optional):
		if set to true, then errors from the invoked Lambda will cause the processor to fail
		defaults to false
*/
type LambdaOptions struct {
	Function       string `json:"function"`
	ErrorOnFailure bool   `json:"error_on_failure"`
}

// ApplyBatch processes a slice of encapsulated data with the Lambda processor. Conditions are optionally applied to the data to enable processing.
func (p Lambda) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process lambda: %v", err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("process lambda: %v", err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Lambda processor.
func (p Lambda) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Function == "" {
		return cap, fmt.Errorf("process lambda: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return cap, fmt.Errorf("process lambda: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
	}

	// lazy load API
	if !lambdaAPI.IsEnabled() {
		lambdaAPI.Setup()
	}

	result := cap.Get(p.InputKey)
	if !result.IsObject() {
		return cap, fmt.Errorf("process lambda: inputkey %s: %v", p.InputKey, errLambdaInputNotAnObject)
	}

	resp, err := lambdaAPI.Invoke(ctx, p.Options.Function, []byte(result.Raw))
	if err != nil {
		return cap, fmt.Errorf("process lambda: %v", err)
	}

	if resp.FunctionError != nil && p.Options.ErrorOnFailure {
		resErr := json.Get(resp.Payload, "errorMessage").String()
		return cap, fmt.Errorf("process lambda: %v", resErr)
	}

	if resp.FunctionError != nil {
		return cap, nil
	}

	if err := cap.Set(p.OutputKey, resp.Payload); err != nil {
		return cap, fmt.Errorf("process lambda: %v", err)
	}

	return cap, nil
}
