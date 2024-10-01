package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &numberBitwiseOR{}

var numberBitwiseORTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": -1,
			},
		},
		[]byte(`0`),
		true,
	},
}

func TestNumberBitwiseOR(t *testing.T) {
	ctx := context.TODO()

	for _, test := range numberBitwiseORTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNumberBitwiseOR(ctx, test.cfg)
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

func benchmarkNumberBitwiseOR(b *testing.B, insp *numberBitwiseOR, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkNumberBitwiseOR(b *testing.B) {
	for _, test := range numberBitwiseORTests {
		insp, err := newNumberBitwiseOR(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNumberBitwiseOR(b, insp, message)
			},
		)
	}
}

func FuzzTestNumberBitwiseOR(f *testing.F) {
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
		insp, err := newNumberBitwiseOR(ctx, config.Config{
			Settings: map[string]interface{}{
				"value": 0x0001,
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
