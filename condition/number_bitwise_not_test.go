package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var numberBitwiseNOTTests = []struct {
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
					"value": "",
				},
			},
		},
		[]byte(`570506001`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"value": "",
				},
			},
		},
		[]byte(`123456789`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"value": "",
				},
			},
		},
		[]byte(`0`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"value": "",
				},
			},
		},
		[]byte(`-1`),
		false,
	},
}

func TestNumberBitwiseNOT(t *testing.T) {
	ctx := context.TODO()

	for _, test := range numberBitwiseNOTTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNumberBitwiseNOT(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Condition(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.test))
			}
		})
	}
}

func benchmarkNumberBitwiseNOT(b *testing.B, insp *numberBitwiseNOT, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkNumberBitwiseNOT(b *testing.B) {
	for _, test := range numberBitwiseNOTTests {
		insp, err := newNumberBitwiseNOT(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNumberBitwiseNOT(b, insp, message)
			},
		)
	}
}

func FuzzTestNumberBitwiseNOT(f *testing.F) {
	testcases := [][]byte{
		[]byte(`570506001`),
		[]byte(`123456789`),
		[]byte(`0`),
		[]byte(`-1`),
		[]byte(`18446744073709551615`), // Max uint64 value
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newNumberBitwiseNOT(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "",
				},
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
