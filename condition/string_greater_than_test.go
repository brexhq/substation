package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &stringGreaterThan{}

var stringGreaterThanTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": "a",
			},
		},
		[]byte("b"),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": "2022-01-01T00:00:00Z",
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
		[]byte(`{"foo":"2023-01-01T00:00:00Z", "bar":"2022-01-01T00:00:00Z"}`),
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
				"value": "greetings",
			},
		},
		[]byte(`{"foo":"hello", "bar":"world"}`),
		false,
	},
}

func TestStringGreaterThan(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringGreaterThanTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newStringGreaterThan(ctx, test.cfg)
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

func benchmarkStringGreaterThan(b *testing.B, insp *stringGreaterThan, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkStringGreaterThan(b *testing.B) {
	for _, test := range stringGreaterThanTests {
		insp, err := newStringGreaterThan(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.data)
				benchmarkStringGreaterThan(b, insp, message)
			},
		)
	}
}

func FuzzTestStringGreaterThan(f *testing.F) {
	testcases := [][]byte{
		[]byte(`"b"`),
		[]byte(`"2023-01-01T00:00:00Z"`),
		[]byte(`{"foo":"2023-01-01T00:00:00Z", "bar":"2022-01-01T00:00:00Z"}`),
		[]byte(`"a"`),
		[]byte(`"z"`),
		[]byte(`" "`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newStringGreaterThan(ctx, config.Config{
			Settings: map[string]interface{}{
				"value": "a",
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
