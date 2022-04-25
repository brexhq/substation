package condition

import (
	"testing"
)

var configTests = []struct {
	conf     []InspectorConfig
	test     []byte
	expected bool
}{
	{
		[]InspectorConfig{
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
		[]InspectorConfig{
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
		[]InspectorConfig{
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
		[]InspectorConfig{
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
		conf     []InspectorConfig
		test     []byte
		expected bool
	}{
		{
			[]InspectorConfig{
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
			[]InspectorConfig{
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
		conf     []InspectorConfig
		test     []byte
		expected bool
	}{
		{
			[]InspectorConfig{
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
			[]InspectorConfig{
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
