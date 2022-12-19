package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	ilambda "github.com/brexhq/substation/internal/aws/lambda"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

var lambdaAPI ilambda.API

// errlambdaInputNotAnObject is returned when the input is not a JSON object.
const errlambdaInputNotAnObject = errors.Error("input is not an object")

type lambda struct {
	process
	Options lambdaOptions `json:"options"`
}

type lambdaOptions struct {
	FunctionName   string `json:"function_name"`
	ErrorOnFailure bool   `json:"error_on_failure"`
}

// Close closes resources opened by the lambda processor.
func (p lambda) Close(context.Context) error {
	return nil
}

func (p lambda) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)

	if err != nil {
		return nil, fmt.Errorf("process lambda: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the lambda processor.
func (p lambda) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.FunctionName == "" {
		return capsule, fmt.Errorf("process lambda: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// only supports JSON, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process lambda: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	// lazy load API
	if !lambdaAPI.IsEnabled() {
		lambdaAPI.Setup()
	}

	result := capsule.Get(p.Key)
	if !result.IsObject() {
		return capsule, fmt.Errorf("process lambda: inputkey %s: %v", p.Key, errlambdaInputNotAnObject)
	}

	resp, err := lambdaAPI.Invoke(ctx, p.Options.FunctionName, []byte(result.Raw))
	if err != nil {
		return capsule, fmt.Errorf("process lambda: %v", err)
	}

	if resp.FunctionError != nil && p.Options.ErrorOnFailure {
		resErr := json.Get(resp.Payload, "errorMessage").String()
		return capsule, fmt.Errorf("process lambda: %v", resErr)
	}

	if resp.FunctionError != nil {
		return capsule, nil
	}

	if err := capsule.Set(p.SetKey, resp.Payload); err != nil {
		return capsule, fmt.Errorf("process lambda: %v", err)
	}

	return capsule, nil
}
