package process

import (
	"context"
	"errors"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/aws/lambda"
	"github.com/brexhq/substation/internal/json"
)

/*
LambdaInput contain custom options settings for this processor.

Payload: maps values from a JSON object (Key) to values in the AWS Lambda payload (PayloadKey)
*/
type LambdaInput struct {
	Payload []struct {
		Key        string `mapstructure:"key"`
		PayloadKey string `mapstructure:"payload_key"`
	} `mapstructure:"payload"`
}

/*
LambdaOptions contain custom options settings for this processor.

Function: the name of the AWS Lambda function to invoke.
Errors: if true, then errors from the invoked Lambda will cause this processor to fail (defaults to false).
*/
type LambdaOptions struct {
	Function string `mapstructure:"function"`
	Errors   bool   `mapstructure:"errors"`
}

// Lambda implements the Byter and Channeler interfaces and synchronously invokes an AWS Lambda. More information is available in the README.
type Lambda struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     LambdaInput              `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   LambdaOptions            `mapstructure:"options"`
}

var lambdaAPI lambda.API

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Lambda) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	var array [][]byte

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	for data := range ch {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			array = append(array, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
		array = append(array, processed)
	}

	output := make(chan []byte, len(array))
	for _, x := range array {
		output <- x
	}
	close(output)
	return output, nil

}

// Byte processes a byte slice with this processor
func (p Lambda) Byte(ctx context.Context, data []byte) ([]byte, error) {
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

	if resp.FunctionError != nil && p.Options.Errors {
		resErr := json.Get(resp.Payload, "errorMessage").String()
		return nil, errors.New(resErr)
	}
	if resp.FunctionError != nil {
		return data, nil
	}

	return json.SetRaw(data, p.Output.Key, resp.Payload)
}
