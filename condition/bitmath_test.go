package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var _ Inspector = inspBitmath{}

var bitmathTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass xor",
		config.Config{
			Type: "bitmath",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "xor",
					"value": 3,
				},
			},
		},
		[]byte(`{"foo":"0"}`),
		true,
	},
	{
		"fail xor",
		config.Config{
			Type: "bitmath",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "xor",
					"value": 42,
				},
			},
		},
		[]byte(`{"foo":"42"}`),
		false,
	},
	{
		"!fail xor",
		config.Config{
			Type: "bitmath",
			Settings: map[string]interface{}{
				"key":    "foo",
				"negate": true,
				"options": map[string]interface{}{
					"type":  "xor",
					"value": 42,
				},
			},
		},
		[]byte(`{"foo":"42"}`),
		true,
	},
	{
		"pass or",
		config.Config{
			Type: "bitmath",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "or",
					"value": -1,
				},
			},
		},
		[]byte(`{"foo":"0"}`),
		true,
	},
	{
		"!pass or",
		config.Config{
			Type: "bitmath",
			Settings: map[string]interface{}{
				"key":    "foo",
				"negate": true,
				"options": map[string]interface{}{
					"type":  "or",
					"value": 1,
				},
			},
		},
		[]byte(`{"foo":"0"}`),
		false,
	},
	{
		"pass and",
		config.Config{
			Type: "bitmath",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "and",
					"value": 0x0001,
				},
			},
		},
		[]byte(`{"foo":"570506001"}`),
		true,
	},
	{
		"fail and",
		config.Config{
			Type: "bitmath",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "and",
					"value": 0x0002,
				},
			},
		},
		[]byte(`{"foo":"570506001"}`),
		false,
	},
	{
		"pass data",
		config.Config{
			Type: "bitmath",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":  "or",
					"value": 1,
				},
			},
		},
		[]byte(`0001`),
		true,
	},
}

func TestBitmath(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range bitmathTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			insp, err := newInspBitmath(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v", test.expected, check)
			}
		})
	}
}

func benchmarkBitmathByte(b *testing.B, inspector inspBitmath, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkBitmathByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range bitmathTests {
		insp, err := newInspBitmath(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkBitmathByte(b, insp, capsule)
			},
		)
	}
}
