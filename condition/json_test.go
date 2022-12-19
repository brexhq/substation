package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var jsonSchemaTests = []struct {
	name      string
	inspector jsonSchema
	test      []byte
	expected  bool
}{
	{
		"string",
		jsonSchema{
			Options: jsonSchemaOptions{
				Schema: []struct {
					Key  string `json:"key"`
					Type string `json:"type"`
				}{
					{Key: "hello", Type: "String"},
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"!string",
		jsonSchema{
			Options: jsonSchemaOptions{
				Schema: []struct {
					Key  string `json:"key"`
					Type string `json:"type"`
				}{
					{Key: "foo", Type: "String"},
				},
			},
		},
		[]byte(`{"foo":123}`),
		false,
	},
	{
		"string array",
		jsonSchema{
			condition: condition{
				Negate: true,
			},
			Options: jsonSchemaOptions{
				Schema: []struct {
					Key  string `json:"key"`
					Type string `json:"type"`
				}{
					{Key: "foo", Type: "String/Array"},
				},
			},
		},
		[]byte(`{"foo":["bar","baz"]}`),
		true,
	},
}

func TestJSONSchema(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range jsonSchemaTests {
		capsule.SetData(test.test)

		check, err := test.inspector.Inspect(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if test.expected != check {
			t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.test))
		}
	}
}

func benchmarkJSONSchemaByte(b *testing.B, inspector jsonSchema, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkJSONSchemaByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range jsonSchemaTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkJSONSchemaByte(b, test.inspector, capsule)
			},
		)
	}
}

var jsonValidTests = []struct {
	name      string
	inspector jsonValid
	test      []byte
	expected  bool
}{
	{
		"valid",
		jsonValid{},
		[]byte(`{"hello":"world"}`),
		true,
	},
	{
		"invalid",
		jsonValid{},
		[]byte(`{hello:"world"}`),
		false,
	},
	{
		"!invalid",
		jsonValid{
			condition: condition{
				Negate: true,
			},
		},
		[]byte(`{"hello":"world"}`),
		false,
	},
	{
		"!valid",
		jsonValid{
			condition: condition{
				Negate: true,
			},
		},
		[]byte(`{hello:"world"}`),
		true,
	},
}

func TestJSONValid(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range jsonValidTests {
		capsule.SetData(test.test)

		check, err := test.inspector.Inspect(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if test.expected != check {
			t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.test))
		}
	}
}

func benchmarkJSONValidByte(b *testing.B, inspector jsonValid, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkJSONValidByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range jsonValidTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkJSONValidByte(b, test.inspector, capsule)
			},
		)
	}
}
