package condition

import (
	"testing"
)

var jsonSchemaTests = []struct {
	name      string
	inspector JSONSchema
	test      []byte
	expected  bool
}{
	{
		"string",
		JSONSchema{
			Schema: []struct {
				Key  string `mapstructure:"key"`
				Type string `mapstructure:"type"`
			}{
				{Key: "hello", Type: "String"},
			},
			Negate: false,
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"!string",
		JSONSchema{
			Schema: []struct {
				Key  string `mapstructure:"key"`
				Type string `mapstructure:"type"`
			}{
				{Key: "foo", Type: "String"},
			},
			Negate: false,
		},
		[]byte(`{"foo":123}`),
		false,
	},
	{
		"string array",
		JSONSchema{
			Schema: []struct {
				Key  string `mapstructure:"key"`
				Type string `mapstructure:"type"`
			}{
				{Key: "foo", Type: "String/Array"},
			},
			Negate: true,
		},
		[]byte(`{"foo":["bar","baz"]}`),
		true,
	},
}

func TestJSONSchema(t *testing.T) {
	for _, testing := range jsonSchemaTests {
		ok, _ := testing.inspector.Inspect(testing.test)

		if testing.expected != ok {
			t.Logf("expected %v, got %v, %v", testing.expected, ok, string(testing.test))
			t.Fail()
		}
	}
}

func benchmarkJSONSchemaByte(b *testing.B, inspector JSONSchema, test []byte) {
	for i := 0; i < b.N; i++ {
		inspector.Inspect(test)
	}
}

func BenchmarkJSONSchemaByte(b *testing.B) {
	for _, test := range jsonSchemaTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkJSONSchemaByte(b, test.inspector, test.test)
			},
		)
	}
}

var jsonValidTests = []struct {
	name      string
	inspector JSONValid
	test      []byte
	expected  bool
}{
	{
		"valid",
		JSONValid{},
		[]byte(`{"hello":"world"}`),
		true,
	},
	{
		"invalid",
		JSONValid{},
		[]byte(`{hello:"world"}`),
		false,
	},
	{
		"!invalid",
		JSONValid{
			Negate: true,
		},
		[]byte(`{"hello":"world"}`),
		false,
	},
	{
		"!valid",
		JSONValid{
			Negate: true,
		},
		[]byte(`{hello:"world"}`),
		true,
	},
}

func TestJSONValid(t *testing.T) {
	for _, testing := range jsonValidTests {
		ok, _ := testing.inspector.Inspect(testing.test)

		if testing.expected != ok {
			t.Logf("expected %v, got %v, %v", testing.expected, ok, string(testing.test))
			t.Fail()
		}
	}
}

func benchmarkJSONValidByte(b *testing.B, inspector JSONValid, test []byte) {
	for i := 0; i < b.N; i++ {
		inspector.Inspect(test)
	}
}

func BenchmarkJSONValidByte(b *testing.B) {
	for _, test := range jsonValidTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkJSONValidByte(b, test.inspector, test.test)
			},
		)
	}
}
