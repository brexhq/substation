package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ inspector = &numberLengthGreaterThan{}

var numberLengthGreaterThanTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "foo",
				},
				"length": 2,
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"length": 2,
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "foo",
				},
				"length": 3,
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"length": 3,
			},
		},
		[]byte(`bar`),
		false,
	},
}

func TestNumberLengthGreaterThan(t *testing.T) {
	ctx := context.TODO()

	for _, test := range numberLengthGreaterThanTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNumberLengthGreaterThan(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v", test.expected, check)
				t.Errorf("settings: %+v", test.cfg)
				t.Errorf("test: %+v", string(test.test))
			}
		})
	}
}

func benchmarkNumberLengthGreaterThan(b *testing.B, insp *numberLengthGreaterThan, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkNumberLengthGreaterThan(b *testing.B) {
	for _, test := range numberLengthGreaterThanTests {
		insp, err := newNumberLengthGreaterThan(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNumberLengthGreaterThan(b, insp, message)
			},
		)
	}
}
