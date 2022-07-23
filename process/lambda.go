package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/aws/lambda"
	"github.com/brexhq/substation/internal/json"
)

/*
LambdaOptions contains custom options settings for the Lambda processor:
	Function:
		function to invoke
	ErrorOnFailure (optional):
		if set to true, then errors from the invoked Lambda will cause the processor to fail
		defaults to false
*/
type LambdaOptions struct {
	Function       string `json:"function"`
	ErrorOnFailure bool   `json:"error_on_failure"`
}

/*
Lambda processes data by synchronously invoking an AWS Lambda and returning the payload. The average latency of synchronously invoking a Lambda function is 10s of milliseconds, but latency can take 100s to 1000s of milliseconds depending on the function which can have significant impact on total event latency. If Substation is running in AWS Lambda with Kinesis, then this latency can be mitigated by increasing the parallelization factor of the Lambda (https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html).

The input key's value must be a JSON object that contains settings for the Lambda. It is recommended to use the copy and insert processors to create the JSON object before calling this processor and to use the delete processor to remove the JSON object after calling this processor.

The processor supports these patterns:
	JSON:
		{"foo":"bar","lambda":{"lookup":"baz"}} >>> {"foo":"bar","lambda":{"baz":"qux"}}

The processor uses this Jsonnet configuration:
	{
		type: 'lambda',
		settings: {
			options: {
				function: 'foo-function',
			},
			input_key: 'lambda',
			output_key: 'lambda',
		},
	}
*/
type Lambda struct {
	Options   LambdaOptions            `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

var lambdaAPI lambda.API

// Slice processes a slice of bytes with the Lambda processor. Conditions are optionally applied on the bytes to enable processing.
func (p Lambda) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	// lazy load API
	if !lambdaAPI.IsEnabled() {
		lambdaAPI.Setup()
	}

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("slicer: %v", err)
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

// Byte processes bytes with the Lambda processor.
func (p Lambda) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Function == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey == "" || p.OutputKey == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// lazy load API
	if !lambdaAPI.IsEnabled() {
		lambdaAPI.Setup()
	}

	payload := json.Get(data, p.InputKey)
	if !payload.IsObject() {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	resp, err := lambdaAPI.Invoke(ctx, p.Options.Function, []byte(payload.Raw))
	if err != nil {
		return nil, fmt.Errorf("byter settings %+v: %w", p, err)
	}

	if resp.FunctionError != nil && p.Options.ErrorOnFailure {
		resErr := json.Get(resp.Payload, "errorMessage").String()
		return nil, fmt.Errorf("byter settings %+v: %v", p, resErr)
	}

	if resp.FunctionError != nil {
		return data, nil
	}

	return json.Set(data, p.OutputKey, resp.Payload)
}
