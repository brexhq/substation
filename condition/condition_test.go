package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

// all inspectors must return true for AND to return true
var conditionANDTests = []struct {
	name     string
	conf     []config.Config
	test     []byte
	expected bool
}{
	{
		"strings",
		[]config.Config{
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "equals",
					"expression": "foo",
				},
			},
		},
		[]byte("foo"),
		true,
	},
	{
		"regexp",
		[]config.Config{
			{
				Type: "regexp",
				Settings: map[string]interface{}{
					"expression": "^foo$",
				},
			},
		},
		[]byte("foo"),
		true,
	},
	{
		"content",
		[]config.Config{
			{
				Type: "content",
				Settings: map[string]interface{}{
					"type": "application/x-gzip",
				},
			},
		},
		[]byte{80, 75, 3, 4},
		false,
	},
	{
		"length",
		[]config.Config{
			{
				Type: "length",
				Settings: map[string]interface{}{
					"value":    3,
					"function": "equals",
				},
			},
		},
		[]byte("foo"),
		true,
	},
	{
		"string length",
		[]config.Config{
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "equals",
					"expression": "foo",
				},
			},
			{
				Type: "length",
				Settings: map[string]interface{}{
					"value":    3,
					"function": "equals",
				},
			},
		},
		[]byte("foo"),
		true,
	},
}

func TestAND(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range conditionANDTests {
		capsule.SetData(test.test)

		cfg := Config{
			Operator:   "and",
			Inspectors: test.conf,
		}

		op, err := OperatorFactory(cfg)
		if err != nil {
			t.Error(err)
		}

		ok, err := op.Operate(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if ok != test.expected {
			t.Errorf("expected %v, got %v", test.expected, ok)
		}
	}
}

func benchmarkAND(b *testing.B, conf []config.Config, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := MakeInspectors(conf)
		op := AND{inspectors}
		_, _ = op.Operate(ctx, capsule)
	}
}

func BenchmarkAND(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range conditionANDTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkAND(b, test.conf, capsule)
			},
		)
	}
}

// any inspector must return true for OR to return true
var conditionORTests = []struct {
	name     string
	conf     []config.Config
	test     []byte
	expected bool
}{
	{
		"strings",
		[]config.Config{
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "equals",
					"expression": "foo",
				},
			},
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "equals",
					"expression": "baz",
				},
			},
		},
		[]byte("foo"),
		true,
	},
	{
		"length",
		[]config.Config{
			{
				Type: "length",
				Settings: map[string]interface{}{
					"value":    3,
					"function": "equals",
				},
			},
			{
				Type: "length",
				Settings: map[string]interface{}{
					"value":    4,
					"function": "equals",
				},
			},
			{
				Type: "length",
				Settings: map[string]interface{}{
					"value":    5,
					"function": "equals",
				},
			},
		},
		[]byte("foo"),
		true,
	},
	{
		"string length",
		[]config.Config{
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "equals",
					"expression": "foo",
				},
			},
			{
				Type: "length",
				Settings: map[string]interface{}{
					"value":    4,
					"function": "equals",
				},
			},
		},
		[]byte("foo"),
		true,
	},
}

func TestOR(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range conditionORTests {
		capsule.SetData(test.test)

		cfg := Config{
			Operator:   "or",
			Inspectors: test.conf,
		}

		op, err := OperatorFactory(cfg)
		if err != nil {
			t.Error(err)
		}

		ok, err := op.Operate(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if ok != test.expected {
			t.Errorf("expected %v, got %v", test.expected, ok)
		}
	}
}

func benchmarkOR(b *testing.B, conf []config.Config, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := MakeInspectors(conf)
		op := OR{inspectors}
		_, _ = op.Operate(ctx, capsule)
	}
}

func BenchmarkOR(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range conditionORTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkOR(b, test.conf, capsule)
			},
		)
	}
}

// all inspectors must return true for NAND to return false
var conditionNANDTests = []struct {
	name     string
	conf     []config.Config
	test     []byte
	expected bool
}{
	{
		"strings",
		[]config.Config{
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "equals",
					"expression": "baz",
				},
			},
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "equals",
					"expression": "qux",
				},
			},
		},
		[]byte("foo"),
		true,
	},
	{
		"strings",
		[]config.Config{
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "equals",
					"expression": "foo",
				},
			},
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "startswith",
					"expression": "f",
				},
			},
		},
		[]byte("foo"),
		false,
	},
}

func TestNAND(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range conditionNANDTests {
		capsule.SetData(test.test)

		cfg := Config{
			Operator:   "nand",
			Inspectors: test.conf,
		}

		op, err := OperatorFactory(cfg)
		if err != nil {
			t.Error(err)
		}

		ok, err := op.Operate(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if ok != test.expected {
			t.Errorf("expected %v, got %v", test.expected, ok)
		}
	}
}

func benchmarkNAND(b *testing.B, conf []config.Config, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := MakeInspectors(conf)
		op := NAND{inspectors}
		_, _ = op.Operate(ctx, capsule)
	}
}

func BenchmarkNAND(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range conditionNORTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkNAND(b, test.conf, capsule)
			},
		)
	}
}

// any inspector must return true for NOR to return false
var conditionNORTests = []struct {
	name     string
	conf     []config.Config
	test     []byte
	expected bool
}{
	{
		"strings",
		[]config.Config{
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "equals",
					"expression": "baz",
				},
			},
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "startswith",
					"expression": "b",
				},
			},
		},
		[]byte("foo"),
		true,
	},
	{
		"strings",
		[]config.Config{
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "equals",
					"expression": "foo",
				},
			},
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "startswith",
					"expression": "b",
				},
			},
		},
		[]byte("foo"),
		false,
	},
}

func TestNOR(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range conditionNORTests {
		capsule.SetData(test.test)

		cfg := Config{
			Operator:   "nor",
			Inspectors: test.conf,
		}

		op, err := OperatorFactory(cfg)
		if err != nil {
			t.Error(err)
		}

		ok, err := op.Operate(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if ok != test.expected {
			t.Errorf("expected %v, got %v", test.expected, ok)
		}
	}
}

func benchmarkNOR(b *testing.B, conf []config.Config, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := MakeInspectors(conf)
		op := NOR{inspectors}
		_, _ = op.Operate(ctx, capsule)
	}
}

func BenchmarkNOR(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range conditionNORTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkNOR(b, test.conf, capsule)
			},
		)
	}
}

func TestFactory(t *testing.T) {
	for _, test := range conditionANDTests {
		_, err := InspectorFactory(test.conf[0])
		if err != nil {
			t.Error(err)
		}
	}
}

func benchmarkFactory(b *testing.B, conf config.Config) {
	for i := 0; i < b.N; i++ {
		_, _ = InspectorFactory(conf)
	}
}

func BenchmarkFactory(b *testing.B) {
	for _, test := range conditionANDTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkFactory(b, test.conf[0])
			},
		)
	}
}
