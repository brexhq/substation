package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

// all inspectors must return true for and to return true
var andTests = []struct {
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
					"options": map[string]interface{}{
						"type":       "equals",
						"expression": "foo",
					},
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
					"options": map[string]interface{}{
						"expression": "^foo$",
					},
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
					"options": map[string]interface{}{
						"type": "application/x-gzip",
					},
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
					"options": map[string]interface{}{
						"value": 3,
						"type":  "equals",
					},
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
					"options": map[string]interface{}{
						"type":       "equals",
						"expression": "foo",
					},
				},
			},
			{
				Type: "length",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"value": 3,
						"type":  "equals",
					},
				},
			},
		},
		[]byte("foo"),
		true,
	},
}

func TestAnd(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range andTests {
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

func benchmarkAnd(b *testing.B, conf []config.Config, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := MakeInspectors(conf...)
		op := and{inspectors}
		_, _ = op.Operate(ctx, capsule)
	}
}

func BenchmarkAnd(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range andTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkAnd(b, test.conf, capsule)
			},
		)
	}
}

// any inspector must return true for or to return true
var orTests = []struct {
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
					"options": map[string]interface{}{
						"type":       "equals",
						"expression": "foo",
					},
				},
			},
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"type":       "equals",
						"expression": "baz",
					},
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
					"options": map[string]interface{}{
						"value": 3,
						"type":  "equals",
					},
				},
			},
			{
				Type: "length",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"value": 4,
						"type":  "equals",
					},
				},
			},
			{
				Type: "length",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"value": 5,
						"type":  "equals",
					},
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
					"options": map[string]interface{}{
						"type":       "equals",
						"expression": "foo",
					},
				},
			},
			{
				Type: "length",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"value": 4,
						"type":  "equals",
					},
				},
			},
		},
		[]byte("foo"),
		true,
	},
}

func TestOr(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range orTests {
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

func benchmarkOr(b *testing.B, conf []config.Config, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := MakeInspectors(conf...)
		op := or{inspectors}
		_, _ = op.Operate(ctx, capsule)
	}
}

func BenchmarkOr(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range orTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkOr(b, test.conf, capsule)
			},
		)
	}
}

// all inspectors must return true for nand to return false
var nandTests = []struct {
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
					"options": map[string]interface{}{
						"type":       "equals",
						"expression": "baz",
					},
				},
			},
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"type":       "equals",
						"expression": "qux",
					},
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
					"options": map[string]interface{}{
						"type":       "equals",
						"expression": "foo",
					},
				},
			},
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"type":       "starts_with",
						"expression": "f",
					},
				},
			},
		},
		[]byte("foo"),
		false,
	},
}

func TestNand(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range nandTests {
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

func benchmarkNand(b *testing.B, conf []config.Config, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := MakeInspectors(conf...)
		op := nand{inspectors}
		_, _ = op.Operate(ctx, capsule)
	}
}

func BenchmarkNand(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range norTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkNand(b, test.conf, capsule)
			},
		)
	}
}

// any inspector must return true for nor to return false
var norTests = []struct {
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
					"options": map[string]interface{}{
						"type":       "equals",
						"expression": "baz",
					},
				},
			},
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"type":       "starts_with",
						"expression": "b",
					},
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
					"options": map[string]interface{}{
						"type":       "equals",
						"expression": "foo",
					},
				},
			},
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"type":       "starts_with",
						"expression": "b",
					},
				},
			},
		},
		[]byte("foo"),
		false,
	},
}

func TestNor(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range norTests {
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

func benchmarkNor(b *testing.B, conf []config.Config, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := MakeInspectors(conf...)
		op := nor{inspectors}
		_, _ = op.Operate(ctx, capsule)
	}
}

func BenchmarkNor(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range norTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkNor(b, test.conf, capsule)
			},
		)
	}
}

func TestFactory(t *testing.T) {
	for _, test := range andTests {
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
	for _, test := range andTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkFactory(b, test.conf[0])
			},
		)
	}
}
