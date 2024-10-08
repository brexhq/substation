package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &stringLessThan{}

var stringLessThanTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": "b",
			},
		},
		[]byte("a"),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": "2024-01",
			},
		},
		[]byte(`2023-01-01T00:00:00Z`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
					"target_key": "bar",
				},
			},
		},
		[]byte(`{"foo":"2022-01-01T00:00:00Z", "bar":"2023-01-01T00:00:00Z"}`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
					"target_key": "bar",
				},
				"value": "2025-01-01",
			},
		},
		[]byte(`{"foo":"2024-01-01T00:00:00Z", "bar":"2023-01-01"}`),
		false,
	},
}

func TestStringLessThan(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringLessThanTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newStringLessThan(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Condition(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v", test.expected, check)
			}
		})
	}
}

func benchmarkStringLessThan(b *testing.B, insp *stringLessThan, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkStringLessThan(b *testing.B) {
	for _, test := range stringLessThanTests {
		insp, err := newStringLessThan(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.data)
				benchmarkStringLessThan(b, insp, message)
			},
		)
	}
}

func FuzzTestStringLessThan(f *testing.F) {
	testcases := [][]byte{
		[]byte(`"a"`),
		[]byte(`"2022-01-01T00:00:00Z"`),
		[]byte(`{"foo":"2022-01-01T00:00:00Z", "bar":"2023-01-01T00:00:00Z"}`),
		[]byte(`"b"`),
		[]byte(`" "`),
		[]byte(`"z"`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newStringLessThan(ctx, config.Config{
			Settings: map[string]interface{}{
				"value": "b",
			},
		})
		if err != nil {
			return
		}

		_, err = insp.Condition(ctx, message)
		if err != nil {
			return
		}
	})
}
