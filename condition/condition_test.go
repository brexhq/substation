package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var allTests = []struct {
	name     string
	conf     []config.Config
	test     []byte
	expected bool
}{
	{
		"format_mime",
		[]config.Config{
			{
				Type: "format_mime",
				Settings: map[string]interface{}{
					"type": "application/x-gzip",
				},
			},
		},
		[]byte{80, 75, 3, 4},
		false,
	},
	{
		"string",
		[]config.Config{
			{
				Type: "string_equal_to",
				Settings: map[string]interface{}{
					"string": "foo",
				},
			},
		},
		[]byte("foo"),
		true,
	},
	{
		"pattern",
		[]config.Config{
			{
				Type: "string_pattern",
				Settings: map[string]interface{}{
					"pattern": "^foo$",
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
				Type: "format_mime",
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
				Type: "number_length_equal_to",
				Settings: map[string]interface{}{
					"length": 3,
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
				Type: "string_equal_to",
				Settings: map[string]interface{}{
					"string": "foo",
				},
			},
			{
				Type: "number_length_equal_to",
				Settings: map[string]interface{}{
					"length": 3,
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
								Type: "string_starts_with",
								Settings: map[string]interface{}{
									"string": "f",
								},
							},
							{
								Type: "string_ends_with",
								Settings: map[string]interface{}{
									"string": "b",
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
								Type: "number_length_equal_to",
								Settings: map[string]interface{}{
									"length": 3,
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
			message := message.New().SetData(test.test)
			cfg := Config{
				Operator:   "all",
				Inspectors: test.conf,
			}

			op, err := New(ctx, cfg)
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

func benchmarkAll(b *testing.B, conf []config.Config, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := newInspectors(ctx, conf...)
		op := opAll{inspectors}
		_, _ = op.Operate(ctx, message)
	}
}

func BenchmarkAll(b *testing.B) {
	for _, test := range allTests {
		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
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
		"string",
		[]config.Config{
			{
				Type: "string_equal_to",
				Settings: map[string]interface{}{
					"string": "foo",
				},
			},
			{
				Type: "string_equal_to",
				Settings: map[string]interface{}{
					"string": "baz",
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
				Type: "number_length_equal_to",
				Settings: map[string]interface{}{
					"length": 3,
				},
			},
			{
				Type: "number_length_equal_to",
				Settings: map[string]interface{}{
					"length": 4,
				},
			},
			{
				Type: "number_length_equal_to",
				Settings: map[string]interface{}{
					"length": 5,
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
				Type: "string_equal_to",
				Settings: map[string]interface{}{
					"string": "foo",
				},
			},
			{
				Type: "number_length_equal_to",
				Settings: map[string]interface{}{
					"length": 4,
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
								Type: "number_length_equal_to",
								Settings: map[string]interface{}{
									"length": 4,
								},
							},
							{
								Type: "string_starts_with",
								Settings: map[string]interface{}{
									"string": "f",
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
								Type: "number_length_equal_to",
								Settings: map[string]interface{}{
									"length": 3,
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
		message := message.New().SetData(test.test)

		cfg := Config{
			Operator:   "any",
			Inspectors: test.conf,
		}

		op, err := New(ctx, cfg)
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

func benchmarkAny(b *testing.B, conf []config.Config, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := newInspectors(ctx, conf...)
		op := opAny{inspectors}
		_, _ = op.Operate(ctx, message)
	}
}

func BenchmarkAny(b *testing.B) {
	for _, test := range anyTests {
		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
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
		"string",
		[]config.Config{
			{
				Type: "string_equal_to",
				Settings: map[string]interface{}{
					"string": "baz",
				},
			},
			{
				Type: "string_starts_with",
				Settings: map[string]interface{}{
					"string": "b",
				},
			},
		},
		[]byte("foo"),
		true,
	},
	{
		"string",
		[]config.Config{
			{
				Type: "string_equal_to",
				Settings: map[string]interface{}{
					"string": "foo",
				},
			},
			{
				Type: "string_starts_with",
				Settings: map[string]interface{}{
					"string": "b",
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
				Type: "number_length_equal_to",
				Settings: map[string]interface{}{
					"type":   "equals",
					"length": 0,
				},
			},
			{
				Type: "meta_negate",
				Settings: map[string]interface{}{
					"inspector": map[string]interface{}{
						"type": "string_starts_with",
						"settings": map[string]interface{}{
							"string": "f",
						},
					},
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
			message := message.New().SetData(test.test)
			cfg := Config{
				Operator:   "none",
				Inspectors: test.conf,
			}

			op, err := New(ctx, cfg)
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

func benchmarkNone(b *testing.B, conf []config.Config, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspectors, _ := newInspectors(ctx, conf...)
		op := opNone{inspectors}
		_, _ = op.Operate(ctx, message)
	}
}

func BenchmarkNone(b *testing.B) {
	for _, test := range noneTests {
		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNone(b, test.conf, message)
			},
		)
	}
}

func TestNewInspector(t *testing.T) {
	for _, test := range allTests {
		_, err := newInspector(context.TODO(), test.conf[0])
		if err != nil {
			t.Error(err)
		}
	}
}

func benchmarknewInspector(b *testing.B, conf config.Config) {
	for i := 0; i < b.N; i++ {
		_, _ = newInspector(context.TODO(), conf)
	}
}

func BenchmarkNewInspector(b *testing.B) {
	for _, test := range allTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarknewInspector(b, test.conf[0])
			},
		)
	}
}
