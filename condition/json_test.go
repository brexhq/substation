package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var jsonSchemaTests = []struct {
	name      string
	inspector inspJSONSchema
	test      []byte
	expected  bool
}{
	{
		"string",
		inspJSONSchema{
			Options: inspJSONSchemaOptions{
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
		inspJSONSchema{
			Options: inspJSONSchemaOptions{
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
		inspJSONSchema{
			condition: condition{
				Negate: true,
			},
			Options: inspJSONSchemaOptions{
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
		var _ Inspector = test.inspector

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

func benchmarkJSONSchemaByte(b *testing.B, inspector inspJSONSchema, capsule config.Capsule) {
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
	inspector inspJSONValid
	test      []byte
	expected  bool
}{
	{
		"valid",
		inspJSONValid{},
		[]byte(`{"hello":"world"}`),
		true,
	},
	{
		"invalid",
		inspJSONValid{},
		[]byte(`{hello:"world"}`),
		false,
	},
	{
		"!invalid",
		inspJSONValid{
			condition: condition{
				Negate: true,
			},
		},
		[]byte(`{"hello":"world"}`),
		false,
	},
	{
		"!valid",
		inspJSONValid{
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

func benchmarkJSONValidByte(b *testing.B, inspector inspJSONValid, capsule config.Capsule) {
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
