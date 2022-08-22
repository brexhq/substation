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
		[]byte{80, 75, 03, 04},
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
	cap := config.NewCapsule()
	for _, test := range conditionANDTests {
		cfg := Config{
			Operator:   "and",
			Inspectors: test.conf,
		}

		op, err := OperatorFactory(cfg)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		cap.SetData(test.test)
		ok, err := op.Operate(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if ok != test.expected {
			t.Logf("expected %v, got %v", test.expected, ok)
			t.Fail()
		}
	}
}

func benchmarkAND(b *testing.B, conf []config.Config, cap config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := MakeInspectors(conf)
		op := AND{inspectors}
		op.Operate(ctx, cap)
	}
}

func BenchmarkAND(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range conditionANDTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkAND(b, test.conf, cap)
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
	cap := config.NewCapsule()
	for _, test := range conditionORTests {
		cfg := Config{
			Operator:   "or",
			Inspectors: test.conf,
		}

		op, err := OperatorFactory(cfg)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		cap.SetData(test.test)
		ok, err := op.Operate(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if ok != test.expected {
			t.Logf("expected %v, got %v", test.expected, ok)
			t.Fail()
		}
	}
}

func benchmarkOR(b *testing.B, conf []config.Config, cap config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := MakeInspectors(conf)
		op := OR{inspectors}
		op.Operate(ctx, cap)
	}
}

func BenchmarkOR(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range conditionORTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkOR(b, test.conf, cap)
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
	cap := config.NewCapsule()
	for _, test := range conditionNANDTests {
		cfg := Config{
			Operator:   "nand",
			Inspectors: test.conf,
		}

		op, err := OperatorFactory(cfg)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		cap.SetData(test.test)
		ok, err := op.Operate(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if ok != test.expected {
			t.Logf("expected %v, got %v", test.expected, ok)
			t.Fail()
		}
	}
}

func benchmarkNAND(b *testing.B, conf []config.Config, cap config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := MakeInspectors(conf)
		op := NAND{inspectors}
		op.Operate(ctx, cap)
	}
}

func BenchmarkNAND(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range conditionNORTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkNAND(b, test.conf, cap)
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
	cap := config.NewCapsule()
	for _, test := range conditionNORTests {
		cfg := Config{
			Operator:   "nor",
			Inspectors: test.conf,
		}

		op, err := OperatorFactory(cfg)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		cap.SetData(test.test)
		ok, err := op.Operate(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if ok != test.expected {
			t.Logf("expected %v, got %v", test.expected, ok)
			t.Fail()
		}
	}
}

func benchmarkNOR(b *testing.B, conf []config.Config, cap config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := MakeInspectors(conf)
		op := NOR{inspectors}
		op.Operate(ctx, cap)
	}
}

func BenchmarkNOR(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range conditionNORTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkNOR(b, test.conf, cap)
			},
		)
	}
}

func TestFactory(t *testing.T) {
	for _, test := range conditionANDTests {
		_, err := InspectorFactory(test.conf[0])
		if err != nil {
			t.Log(err)
			t.Fail()
		}
	}
}

func benchmarkFactory(b *testing.B, conf config.Config) {
	for i := 0; i < b.N; i++ {
		InspectorFactory(conf)
	}
}

func BenchmarkFactory(b *testing.B) {
	for _, test := range conditionANDTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkFactory(b, test.conf[0])
			},
		)
	}
}
