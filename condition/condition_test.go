package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var allTests = []struct {
	name     string
	conf     []config.Config
	test     []byte
	expected bool
}{
	{
		"strings",
		[]config.Config{
			{
				Type: "insp_strings",
				Settings: map[string]interface{}{
					"type":       "equals",
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
				Type: "insp_regexp",
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
				Type: "insp_content",
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
				Type: "insp_length",
				Settings: map[string]interface{}{
					"value": 3,
					"type":  "equals",
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
				Type: "insp_strings",
				Settings: map[string]interface{}{
					"type":       "equals",
					"expression": "foo",
				},
			},
			{
				Type: "insp_length",
				Settings: map[string]interface{}{
					"value": 3,
					"type":  "equals",
				},
			},
		},
		[]byte("foo"),
		true,
	},
	// this test joins multiple ANY operators with an ALL operator, implementing the following logic:
	// if ( "foo" starts with "f" OR "foo" ends with "b" ) AND ( len("foo") == 3 ) then return true
	{
		"condition",
		[]config.Config{
			{
				Type: "meta_condition",
				Settings: map[string]interface{}{
					"condition": map[string]interface{}{
						"operator": "any",
						"inspectors": []config.Config{
							{
								Type: "insp_strings",
								Settings: map[string]interface{}{
									"expression": "f",
									"type":       "starts_with",
								},
							},
							{
								Type: "insp_strings",
								Settings: map[string]interface{}{
									"expression": "b",
									"type":       "ends_with",
								},
							},
						},
					},
				},
			},
			{
				Type: "meta_condition",
				Settings: map[string]interface{}{
					"condition": map[string]interface{}{
						"operator": "all",
						"inspectors": []config.Config{
							{
								Type: "insp_length",
								Settings: map[string]interface{}{
									"value": 3,
									"type":  "equals",
								},
							},
						},
					},
				},
			},
		},
		[]byte("foo"),
		true,
	},
}

func TestAll(t *testing.T) {
	ctx := context.TODO()

	for _, test := range allTests {
		t.Run(test.name, func(t *testing.T) {
			message, _ := mess.New(
				mess.SetData(test.test),
			)

			cfg := Config{
				Operator:   "all",
				Inspectors: test.conf,
			}

			op, err := NewOperator(ctx, cfg)
			if err != nil {
				t.Error(err)
			}

			ok, err := op.Operate(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if ok != test.expected {
				t.Errorf("expected %v, got %v", test.expected, ok)
			}
		})
	}
}

func benchmarkAll(b *testing.B, conf []config.Config, message *mess.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := NewInspectors(ctx, conf...)
		op := opAll{inspectors}
		_, _ = op.Operate(ctx, message)
	}
}

func BenchmarkAll(b *testing.B) {
	for _, test := range allTests {
		b.Run(test.name,
			func(b *testing.B) {
				message, _ := mess.New(
					mess.SetData(test.test),
				)
				benchmarkAll(b, test.conf, message)
			},
		)
	}
}

var anyTests = []struct {
	name     string
	conf     []config.Config
	test     []byte
	expected bool
}{
	{
		"strings",
		[]config.Config{
			{
				Type: "insp_strings",
				Settings: map[string]interface{}{
					"type":       "equals",
					"expression": "foo",
				},
			},
			{
				Type: "insp_strings",
				Settings: map[string]interface{}{
					"type":       "equals",
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
				Type: "insp_length",
				Settings: map[string]interface{}{
					"value": 3,
					"type":  "equals",
				},
			},
			{
				Type: "insp_length",
				Settings: map[string]interface{}{
					"value": 4,
					"type":  "equals",
				},
			},
			{
				Type: "insp_length",
				Settings: map[string]interface{}{
					"value": 5,
					"type":  "equals",
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
				Type: "insp_strings",
				Settings: map[string]interface{}{
					"type":       "equals",
					"expression": "foo",
				},
			},
			{
				Type: "insp_length",
				Settings: map[string]interface{}{
					"value": 4,
					"type":  "equals",
				},
			},
		},
		[]byte("foo"),
		true,
	},
	// this test joins multiple ALL operators with an ANY operator, implementing the following logic:
	// if ( len("foo") == 4 AND "foo" starts with "f" ) OR ( len("foo") == 3 ) then return true
	{
		"condition",
		[]config.Config{
			{
				Type: "meta_condition",
				Settings: map[string]interface{}{
					"condition": map[string]interface{}{
						"operator": "all",
						"inspectors": []config.Config{
							{
								Type: "insp_length",
								Settings: map[string]interface{}{
									"value": 4,
									"type":  "equals",
								},
							},
							{
								Type: "insp_strings",
								Settings: map[string]interface{}{
									"expression": "f",
									"type":       "starts_with",
								},
							},
						},
					},
				},
			},
			{
				Type: "meta_condition",
				Settings: map[string]interface{}{
					"condition": map[string]interface{}{
						"operator": "all",
						"inspectors": []config.Config{
							{
								Type: "insp_length",
								Settings: map[string]interface{}{
									"value": 3,
									"type":  "equals",
								},
							},
						},
					},
				},
			},
		},
		[]byte("foo"),
		true,
	},
}

func TestAny(t *testing.T) {
	ctx := context.TODO()

	for _, test := range anyTests {
		message, _ := mess.New(
			mess.SetData(test.test),
		)

		cfg := Config{
			Operator:   "any",
			Inspectors: test.conf,
		}

		op, err := NewOperator(ctx, cfg)
		if err != nil {
			t.Error(err)
		}

		ok, err := op.Operate(ctx, message)
		if err != nil {
			t.Error(err)
		}

		if ok != test.expected {
			t.Errorf("expected %v, got %v", test.expected, ok)
		}
	}
}

func benchmarkAny(b *testing.B, conf []config.Config, message *mess.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := NewInspectors(ctx, conf...)
		op := opAny{inspectors}
		_, _ = op.Operate(ctx, message)
	}
}

func BenchmarkAny(b *testing.B) {
	for _, test := range anyTests {
		b.Run(test.name,
			func(b *testing.B) {
				message, _ := mess.New(
					mess.SetData(test.test),
				)
				benchmarkAny(b, test.conf, message)
			},
		)
	}
}

var noneTests = []struct {
	name     string
	conf     []config.Config
	test     []byte
	expected bool
}{
	{
		"strings",
		[]config.Config{
			{
				Type: "insp_strings",
				Settings: map[string]interface{}{
					"type":       "equals",
					"expression": "baz",
				},
			},
			{
				Type: "insp_strings",
				Settings: map[string]interface{}{
					"type":       "starts_with",
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
				Type: "insp_strings",
				Settings: map[string]interface{}{
					"type":       "equals",
					"expression": "foo",
				},
			},
			{
				Type: "insp_strings",
				Settings: map[string]interface{}{
					"type":       "starts_with",
					"expression": "b",
				},
			},
		},
		[]byte("foo"),
		false,
	},
	{
		"length",
		[]config.Config{
			{
				Type: "insp_length",
				Settings: map[string]interface{}{
					"type":  "equals",
					"value": 0,
				},
			},
			{
				Type: "insp_strings",
				Settings: map[string]interface{}{
					"type":       "starts_with",
					"expression": "f",
					"negate":     true,
				},
			},
		},
		[]byte("foo"),
		true,
	},
}

func TestNone(t *testing.T) {
	ctx := context.TODO()

	for _, test := range noneTests {
		t.Run(test.name, func(t *testing.T) {
			message, _ := mess.New(
				mess.SetData(test.test),
			)

			cfg := Config{
				Operator:   "none",
				Inspectors: test.conf,
			}

			op, err := NewOperator(ctx, cfg)
			if err != nil {
				t.Error(err)
			}

			ok, err := op.Operate(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if ok != test.expected {
				t.Errorf("expected %v, got %v", test.expected, ok)
			}
		})
	}
}

func benchmarkNone(b *testing.B, conf []config.Config, message *mess.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := NewInspectors(ctx, conf...)
		op := opNone{inspectors}
		_, _ = op.Operate(ctx, message)
	}
}

func BenchmarkNone(b *testing.B) {
	for _, test := range noneTests {
		b.Run(test.name,
			func(b *testing.B) {
				message, _ := mess.New(
					mess.SetData(test.test),
				)
				benchmarkNone(b, test.conf, message)
			},
		)
	}
}

func TestNewInspector(t *testing.T) {
	for _, test := range allTests {
		_, err := NewInspector(context.TODO(), test.conf[0])
		if err != nil {
			t.Error(err)
		}
	}
}

func benchmarkNewInspector(b *testing.B, conf config.Config) {
	for i := 0; i < b.N; i++ {
		_, _ = NewInspector(context.TODO(), conf)
	}
}

func BenchmarkNewInspector(b *testing.B) {
	for _, test := range allTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkNewInspector(b, test.conf[0])
			},
		)
	}
}
