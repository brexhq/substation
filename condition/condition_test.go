package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
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
	// this test joins multiple ANY operators with an ALL operator, implementing the following logic:
	// if ( "foo" starts with "f" OR "foo" ends with "b" ) AND ( len("foo") == 3 ) then return true
	{
		"condition",
		[]config.Config{
			{
				Type: "condition",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"operator": "any",
						"inspectors": []config.Config{
							{
								Type: "strings",
								Settings: map[string]interface{}{
									"options": map[string]interface{}{
										"expression": "f",
										"type":       "starts_with",
									},
								},
							},
							{
								Type: "strings",
								Settings: map[string]interface{}{
									"options": map[string]interface{}{
										"expression": "b",
										"type":       "ends_with",
									},
								},
							},
						},
					},
				},
			},
			{
				Type: "condition",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"operator": "all",
						"inspectors": []config.Config{
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
	capsule := config.NewCapsule()

	for _, test := range allTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			cfg := Config{
				Operator:   "all",
				Inspectors: test.conf,
			}

			op, err := NewOperator(cfg)
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
		})
	}
}

func benchmarkAll(b *testing.B, conf []config.Config, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := NewInspectors(conf...)
		op := opAll{inspectors}
		_, _ = op.Operate(ctx, capsule)
	}
}

func BenchmarkAll(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range allTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkAll(b, test.conf, capsule)
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
	// this test joins multiple ALL operators with an ANY operator, implementing the following logic:
	// if ( len("foo") == 4 AND "foo" starts with "f" ) OR ( len("foo") == 3 ) then return true
	{
		"condition",
		[]config.Config{
			{
				Type: "condition",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"operator": "all",
						"inspectors": []config.Config{
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
								Type: "strings",
								Settings: map[string]interface{}{
									"options": map[string]interface{}{
										"expression": "f",
										"type":       "starts_with",
									},
								},
							},
						},
					},
				},
			},
			{
				Type: "condition",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"operator": "all",
						"inspectors": []config.Config{
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
	capsule := config.NewCapsule()

	for _, test := range anyTests {
		capsule.SetData(test.test)

		cfg := Config{
			Operator:   "any",
			Inspectors: test.conf,
		}

		op, err := NewOperator(cfg)
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

func benchmarkAny(b *testing.B, conf []config.Config, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := NewInspectors(conf...)
		op := opAny{inspectors}
		_, _ = op.Operate(ctx, capsule)
	}
}

func BenchmarkAny(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range anyTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkAny(b, test.conf, capsule)
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
	{
		"length",
		[]config.Config{
			{
				Type: "length",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"type":  "equals",
						"value": 0,
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
					"negate": true,
				},
			},
		},
		[]byte("foo"),
		true,
	},
}

func TestNone(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range noneTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			cfg := Config{
				Operator:   "none",
				Inspectors: test.conf,
			}

			op, err := NewOperator(cfg)
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
		})
	}
}

func benchmarkNone(b *testing.B, conf []config.Config, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := NewInspectors(conf...)
		op := opNone{inspectors}
		_, _ = op.Operate(ctx, capsule)
	}
}

func BenchmarkNone(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range noneTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkNone(b, test.conf, capsule)
			},
		)
	}
}

func TestNewInspector(t *testing.T) {
	for _, test := range allTests {
		_, err := NewInspector(test.conf[0])
		if err != nil {
			t.Error(err)
		}
	}
}

func benchmarkNewInspector(b *testing.B, conf config.Config) {
	for i := 0; i < b.N; i++ {
		_, _ = NewInspector(conf)
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

var conditionTests = []struct {
	name      string
	inspector inspCondition
	test      []byte
	expected  bool
}{
	{
		"object",
		inspCondition{
			Options: Config{
				Operator: "all",
				Inspectors: []config.Config{
					{
						Type: "ip",
						Settings: map[string]interface{}{
							"key": "ip_address",
							"options": map[string]interface{}{
								"type": "private",
							},
						},
					},
				},
			},
		},
		[]byte(`{"ip_address":"192.168.1.2"}`),
		true,
	},
	{
		"data",
		inspCondition{
			Options: Config{
				Operator: "all",
				Inspectors: []config.Config{
					{
						Type: "ip",
						Settings: map[string]interface{}{
							"options": map[string]interface{}{
								"type": "private",
							},
						},
					},
				},
			},
		},
		[]byte("192.168.1.2"),
		true,
	},
}

func TestCondition(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range conditionTests {
		t.Run(test.name, func(t *testing.T) {
			var _ Inspector = test.inspector

			capsule.SetData(test.test)

			check, err := test.inspector.Inspect(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.test))
			}
		})
	}
}

func benchmarkCondition(b *testing.B, inspector inspCondition, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkCondition(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range conditionTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkCondition(b, test.inspector, capsule)
			},
		)
	}
}
