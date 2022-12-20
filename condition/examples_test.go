package condition_test

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

func ExampleInspectorFactory() {
	// data must be gzip
	cfg := config.Config{
		Type: "content",
		Settings: map[string]interface{}{
			"options": map[string]interface{}{
				"type": "application/x-gzip",
			},
		},
	}

	// inspector is retrieved from the factory
	inspector, err := condition.InspectorFactory(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(inspector)
}

func ExampleMakeInspectors() {
	// data must be gzip
	cfg := config.Config{
		Type: "content",
		Settings: map[string]interface{}{
			"options": map[string]interface{}{
				"type": "application/x-gzip",
			},
		},
	}

	// one or more inspectors are created
	inspectors, err := condition.MakeInspectors(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	for _, ins := range inspectors {
		fmt.Println(ins)
	}
}

func ExampleInspectBytes() {
	// data must be gzip
	cfg := config.Config{
		Type: "content",
		Settings: map[string]interface{}{
			"options": map[string]interface{}{
				"type": "application/x-gzip",
			},
		},
	}

	// inspector is retrieved from the factory
	inspector, err := condition.InspectorFactory(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// inspector is applied to bytes
	b := []byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255}
	ok, err := condition.InspectBytes(context.TODO(), b, inspector)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(ok)
	// Output: true
}

func ExampleOperatorFactory() {
	// data must have a length greater than zero and contain
	// the substring "iz"
	cfg := []config.Config{
		{
			Type: "length",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":  "greater_than",
					"value": 0,
				},
			},
		},
		{
			Type: "strings",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "contains",
					"expression": "iz",
				},
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
	operator, err := condition.OperatorFactory(opCfg)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(operator)
}

func ExampleOperateBytes() {
	// data must have a length greater than zero and contain
	// the substring "iz"
	cfg := []config.Config{
		{
			Type: "length",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":  "less_than",
					"value": 6,
				},
			},
		},
		{
			Type: "strings",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "contains",
					"expression": "iz",
				},
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
	operator, err := condition.OperatorFactory(opCfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// operator is applied to bytes
	b := []byte("fizzy")
	ok, err := condition.OperateBytes(context.TODO(), b, operator)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(ok)
	// Output: true
}

func Example_inspect() {
	// data must be gzip
	cfg := config.Config{
		Type: "content",
		Settings: map[string]interface{}{
			"options": map[string]interface{}{
				"type": "application/x-gzip",
			},
		},
	}

	// inspector is retrieved from the factory
	inspector, err := condition.InspectorFactory(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// inspector is applied to capsule
	capsule := config.NewCapsule()
	capsule.SetData([]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255})

	ok, err := inspector.Inspect(context.TODO(), capsule)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(ok)
	// Output: true
}

func Example_operate() {
	// data must have a length greater than zero and contain
	// the substring "iz"
	cfg := []config.Config{
		{
			Type: "length",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":  "less_than",
					"value": 6,
				},
			},
		},
		{
			Type: "strings",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "contains",
					"expression": "iz",
				},
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
	operator, err := condition.OperatorFactory(opCfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// operator is applied to capsule
	capsule := config.NewCapsule()
	capsule.SetData([]byte("fizzy"))

	ok, err := operator.Operate(context.TODO(), capsule)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(ok)
	// Output: true
}
