package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &numberLengthLessThan{}

var numberLengthLessThanTests = []struct {
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
					"source_key": "foo",
				},
				"value": 4,
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": 4,
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
					"source_key": "foo",
				},
				"value": 3,
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"value": 3,
			},
		},
		[]byte(`bar`),
		false,
	},
}

func TestNumberLengthLessThan(t *testing.T) {
	ctx := context.TODO()

	for _, test := range numberLengthLessThanTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNumberLengthLessThan(ctx, test.cfg)
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

func benchmarkNumberLengthLessThan(b *testing.B, insp *numberLengthLessThan, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkNumberLengthLessThan(b *testing.B) {
	for _, test := range numberLengthLessThanTests {
		insp, err := newNumberLengthLessThan(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNumberLengthLessThan(b, insp, message)
			},
		)
	}
}

func FuzzTestNumberLengthLessThan(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"foo":"bar"}`),
		[]byte(`bar`),
		[]byte(`{"foo":"bars"}`),
		[]byte(`bars`),
		[]byte(`{"foo":""}`),
		[]byte(`""`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newNumberLengthLessThan(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
				},
				"value": 4,
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
