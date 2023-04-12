package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Inspector = inspJSONSchema{}
	_ Inspector = inspJSONValid{}
)

var jsonSchemaTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"string",
		config.Config{
			Type: "json_schema",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"schema": []struct {
						Key  string `json:"key"`
						Type string `json:"type"`
					}{
						{Key: "hello", Type: "String"},
					},
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"!string",
		config.Config{
			Type: "json_schema",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"schema": []struct {
						Key  string `json:"key"`
						Type string `json:"type"`
					}{
						{Key: "foo", Type: "String"},
					},
				},
			},
		},
		[]byte(`{"foo":123}`),
		false,
	},
	{
		"string array",
		config.Config{
			Type: "json_schema",
			Settings: map[string]interface{}{
				"negate": true,
				"options": map[string]interface{}{
					"schema": []struct {
						Key  string `json:"key"`
						Type string `json:"type"`
					}{
						{Key: "foo", Type: "String/Array"},
					},
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
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			insp, err := newInspJSONSchema(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.test))
			}
		})
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
		insp, err := newInspJSONSchema(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkJSONSchemaByte(b, insp, capsule)
			},
		)
	}
}

var jsonValidTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"valid",
		config.Config{
			Type: "json_valid",
		},
		[]byte(`{"hello":"world"}`),
		true,
	},
	{
		"invalid",
		config.Config{
			Type: "json_valid",
		},
		[]byte(`{hello:"world"}`),
		false,
	},
	{
		"!invalid",
		config.Config{
			Type: "json_valid",
			Settings: map[string]interface{}{
				"negate": true,
			},
		},
		[]byte(`{"hello":"world"}`),
		false,
	},
	{
		"!valid",
		config.Config{
			Type: "json_valid",
			Settings: map[string]interface{}{
				"negate": true,
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
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			insp, err := newInspJSONValid(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.test))
			}
		})
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
		insp, err := newInspJSONValid(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkJSONValidByte(b, insp, capsule)
			},
		)
	}
}
