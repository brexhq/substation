package condition_test

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

func ExampleNew() {
	// data must have a length greater than zero and contain
	// the substring "iz"
	cfg := []config.Config{
		{
			Type: "insp_length",
			Settings: map[string]interface{}{
				"type":  "greater_than",
				"value": 0,
			},
		},
		{
			Type: "insp_string",
			Settings: map[string]interface{}{
				"type":       "contains",
				"expression": "iz",
			},
		},
	}

	// multiple inspectors are paired with an operator to
	// test many conditions at once.
	opCfg := condition.Config{
		Operator:   "and",
		Inspectors: cfg,
	}

	// operators are retrieved from the factory.
	operator, err := condition.New(context.TODO(), opCfg)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(operator)
}

func Example_operate() {
	ctx := context.TODO()
	// data must have a length greater than zero and contain
	// the substring "iz"
	cfg := []config.Config{
		{
			Type: "insp_length",
			Settings: map[string]interface{}{
				"type":  "less_than",
				"value": 6,
			},
		},

		{
			Type: "insp_string",
			Settings: map[string]interface{}{
				"type":       "contains",
				"expression": "iz",
			},
		},
	}

	// multiple inspectors are paired with an operator to
	// test many conditions at once
	opCfg := condition.Config{
		Operator:   "and",
		Inspectors: cfg,
	}

	// operator is retrieved from the factory
	operator, err := condition.New(ctx, opCfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// operator is applied to message
	message, err := mess.New(
		mess.SetData([]byte("fizzy")),
	)
	if err != nil {
		// handle err
		panic(err)
	}

	ok, err := operator.Operate(ctx, message)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(ok)
	// Output: true
}
