package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ inspector = &numberLengthEqualTo{}

var numberLengthEqualToTests = []struct {
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
					"source_key": "a",
				},
				"value": 3,
			},
		},
		[]byte(`{"a":"bcd"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": 3,
			},
		},
		[]byte(`bcd`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
				},
				"value": 4,
			},
		},
		[]byte(`{"a":"bcd"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"value": 4,
			},
		},
		[]byte(`bcd`),
		false,
	},
}

func TestNumberLengthEqualTo(t *testing.T) {
	ctx := context.TODO()

	for _, test := range numberLengthEqualToTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNumberLengthEqualTo(ctx, test.cfg)
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

func benchmarkNumberLengthEqualTo(b *testing.B, insp *numberLengthEqualTo, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkNumberLengthEqualTo(b *testing.B) {
	for _, test := range numberLengthEqualToTests {
		insp, err := newNumberLengthEqualTo(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNumberLengthEqualTo(b, insp, message)
			},
		)
	}
}
