package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/aws/lambda"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// LambdaInvalidSettings is returned when the Lambda processor is configured with invalid Input and Output settings.
const LambdaInvalidSettings = errors.Error("LambdaInvalidSettings")

/*
LambdaInput contains custom input settings for the Lambda processor:
	Payload:
		maps values from the JSON object (Key) to values in the AWS Lambda payload (PayloadKey)
*/
type LambdaInput struct {
	Payload []struct {
		Key        string `json:"key"`
		PayloadKey string `json:"payload_key"`
	} `json:"payload"`
}

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

The processor supports these patterns:
	json:
		{"foo":"bar"} >>> {"foo":"bar","lambda":{"baz":"qux"}}

The processor uses this Jsonnet configuration:
	{
		type: 'lambda',
		settings: {
			input: {
				payload: [
					{
						key: 'foo',
						payload_key: 'foo',
					}
				],
			},
			output: {
				key: 'lambda',
			}
			options: {
				function: 'foo-function',
			}
		},
	}
*/
type Lambda struct {
	Condition condition.OperatorConfig `json:"condition"`
	Input     LambdaInput              `json:"input"`
	Output    Output                   `json:"output"`
	Options   LambdaOptions            `json:"options"`
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
		return nil, err
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

// Byte processes bytes with the Lambda processor.
func (p Lambda) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// only supports json, so error early if there are no keys
	if len(p.Input.Payload) == 0 && p.Output.Key == "" {
		return nil, LambdaInvalidSettings
	}

	// lazy load API
	if !lambdaAPI.IsEnabled() {
		lambdaAPI.Setup()
	}

	var payload []byte
	var err error
	for _, p := range p.Input.Payload {
		v := json.Get(data, p.Key)
		payload, err = json.Set(payload, p.PayloadKey, v)
		if err != nil {
			return nil, err
		}
	}

	resp, err := lambdaAPI.Invoke(ctx, p.Options.Function, payload)
	if err != nil {
		return nil, err
	}

	if resp.FunctionError != nil && p.Options.ErrorOnFailure {
		resErr := json.Get(resp.Payload, "errorMessage").String()
		return nil, errors.Error(resErr)
	}

	if resp.FunctionError != nil {
		return data, nil
	}

	return json.SetRaw(data, p.Output.Key, resp.Payload)
}
