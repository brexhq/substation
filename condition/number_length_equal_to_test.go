package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &numberLengthEqualTo{}

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

			check, err := insp.Condition(ctx, message)
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
		_, _ = insp.Condition(ctx, message)
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

func FuzzTestNumberLengthEqualTo(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":"bcd"}`),
		[]byte(`bcd`),
		[]byte(`{"a":"bcde"}`),
		[]byte(`abcd`),
		[]byte(`{"a":""}`),
		[]byte(`""`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newNumberLengthEqualTo(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
				},
				"value": 3,
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
