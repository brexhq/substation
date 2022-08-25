package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
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
				Key  string `json:"key"`
				Type string `json:"type"`
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
				Key  string `json:"key"`
				Type string `json:"type"`
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
				Key  string `json:"key"`
				Type string `json:"type"`
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
	ctx := context.TODO()
	cap := config.NewCapsule()

	for _, test := range jsonSchemaTests {
		cap.SetData(test.test)

		check, err := test.inspector.Inspect(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if test.expected != check {
			t.Logf("expected %v, got %v, %v", test.expected, check, string(test.test))
			t.Fail()
		}
	}
}

func benchmarkJSONSchemaByte(b *testing.B, inspector JSONSchema, cap config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspector.Inspect(ctx, cap)
	}
}

func BenchmarkJSONSchemaByte(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range jsonSchemaTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkJSONSchemaByte(b, test.inspector, cap)
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
	ctx := context.TODO()
	cap := config.NewCapsule()

	for _, test := range jsonValidTests {
		cap.SetData(test.test)

		check, err := test.inspector.Inspect(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if test.expected != check {
			t.Logf("expected %v, got %v, %v", test.expected, check, string(test.test))
			t.Fail()
		}
	}
}

func benchmarkJSONValidByte(b *testing.B, inspector JSONValid, cap config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspector.Inspect(ctx, cap)
	}
}

func BenchmarkJSONValidByte(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range jsonValidTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkJSONValidByte(b, test.inspector, cap)
			},
		)
	}
}
