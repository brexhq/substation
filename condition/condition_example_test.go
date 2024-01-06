package condition_test

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func ExampleOperator() {
	ctx := context.TODO()

	// Multiple inspectors can be chained together with an operator.
	// This example uses the "all" operator, which requires all inspectors to
	// return true for the operator to return true.
	cfg := condition.Config{
		Operator: "all",
		Inspectors: []config.Config{
			{
				Type: "number_length_less_than",
				Settings: map[string]interface{}{
					"value": 10,
				},
			},
			{
				Type: "string_contains",
				Settings: map[string]interface{}{
					"value": "f",
				},
			},
		},
	}

	// Operators are retrieved from the factory and
	// applied to a message.
	op, err := condition.New(ctx, cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	msg := message.New().SetData([]byte("fizzy"))
	if err != nil {
		// handle err
		panic(err)
	}

	ok, err := op.Operate(ctx, msg)
	if err != nil {
		// handle err
		panic(err)
	}

	// Output: true
	fmt.Println(ok)
}
