package condition

import (
	"testing"

	"github.com/brexhq/substation/config"
)

var configTests = []struct {
	conf     []config.Config
	test     []byte
	expected bool
}{
	{
		[]config.Config{
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "startswith",
					"expression": "Test",
				},
			},
			{
				Type: "regexp",
				Settings: map[string]interface{}{
					"expression": "^Test$",
				},
			},
		},
		[]byte("Test"),
		true,
	},
	{
		[]config.Config{
			{
				Type: "regexp",
				Settings: map[string]interface{}{
					"expression": "^Tester",
				},
			},
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "startswith",
					"expression": "Test",
				},
			},
		},
		[]byte("Test"),
		false,
	},
	{
		[]config.Config{
			{
				Type: "strings",
				Settings: map[string]interface{}{
					"function":   "equals",
					"expression": "",
					"key":        "baz",
					"negate":     true,
				},
			},
		},
		[]byte(`{"foo":"bar","baz":"hello"`),
		true,
	},
	{
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
}

func TestAND(t *testing.T) {
	for _, test := range configTests {
		cfg := OperatorConfig{
			Operator:   "and",
			Inspectors: test.conf,
		}

		op, err := OperatorFactory(cfg)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		ok, err := op.Operate(test.test)
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

func TestFactory(t *testing.T) {
	for _, test := range configTests {
		conf := test.conf[0]
		c, err := InspectorFactory(conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		check, err := c.Inspect(test.test)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if check != test.expected {
			t.Logf("expected %v, got %v", test.expected, check)
			t.Fail()
		}
	}
}

func TestOR(t *testing.T) {
	var tests = []struct {
		conf     []config.Config
		test     []byte
		expected bool
	}{
		{
			[]config.Config{
				{
					Type: "strings",
					Settings: map[string]interface{}{
						"function":   "startswith",
						"expression": "Test",
					},
				},
				{
					Type: "regexp",
					Settings: map[string]interface{}{
						"expression": "^Test$",
					},
				},
			},
			[]byte("Test"),
			true,
		},
		{
			[]config.Config{
				{
					Type: "regexp",
					Settings: map[string]interface{}{
						"expression": "^Tester",
					},
				},
				{
					Type: "strings",
					Settings: map[string]interface{}{
						"function":   "startswith",
						"expression": "Test",
					},
				},
			},
			[]byte("Test"),
			true,
		},
	}

	for _, test := range tests {
		cfg := OperatorConfig{
			Operator:   "or",
			Inspectors: test.conf,
		}

		op, err := OperatorFactory(cfg)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		ok, err := op.Operate(test.test)
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

func TestNOR(t *testing.T) {
	var tests = []struct {
		conf     []config.Config
		test     []byte
		expected bool
	}{
		{
			[]config.Config{
				{
					Type: "strings",
					Settings: map[string]interface{}{
						"function":   "startswith",
						"expression": "Test",
					},
				},
				{
					Type: "regexp",
					Settings: map[string]interface{}{
						"expression": "^Test$",
					},
				},
			},
			[]byte("Test"),
			false,
		},
		{
			[]config.Config{
				{
					Type: "regexp",
					Settings: map[string]interface{}{
						"expression": "^Tester",
					},
				},
				{
					Type: "strings",
					Settings: map[string]interface{}{
						"function":   "startswith",
						"expression": "Test",
					},
				},
			},
			[]byte("Test"),
			false,
		},
	}

	for _, test := range tests {
		cfg := OperatorConfig{
			Operator:   "nor",
			Inspectors: test.conf,
		}

		op, err := OperatorFactory(cfg)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		ok, err := op.Operate(test.test)
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

var BenchmarkCfg = OperatorConfig{
	Operator:   "and",
	Inspectors: configTests[0].conf,
}

func BenchmarkOperatorFactory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		op, err := OperatorFactory(BenchmarkCfg)
		if err != nil {
			b.Log(err)
			b.Fail()
		}

		op.Operate(configTests[0].test)
	}
}

var op, _ = OperatorFactory(BenchmarkCfg)

func BenchmarkOperate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		op.Operate(configTests[0].test)
	}
}
