package condition_test

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func ExampleNew() {
	// data must have a length greater than zero and contain
	// the substring "iz"
	cfg := []config.Config{
		{
			Type: "logic_len_greater_than",
			Settings: map[string]interface{}{
				"length": 0,
			},
		},
		{
			Type: "string_contains",
			Settings: map[string]interface{}{
				"string": "iz",
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
			Type: "logic_len_less_than",
			Settings: map[string]interface{}{
				"length": 6,
			},
		},

		{
			Type: "string_contains",
			Settings: map[string]interface{}{
				"string": "iz",
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
	msg := message.New().SetData([]byte("fizzy"))
	if err != nil {
		// handle err
		panic(err)
	}

	ok, err := operator.Operate(ctx, msg)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(ok)
	// Output: true
}
