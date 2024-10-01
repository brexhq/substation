package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &stringMatch{}

var stringMatchTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"pattern": "^Test",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"pattern": "^Test",
			},
		},
		[]byte("-Test"),
		false,
	},
}

func TestStringMatch(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringMatchTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newStringMatch(ctx, test.cfg)
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

func benchmarkStringMatchByte(b *testing.B, insp *stringMatch, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkStringMatchByte(b *testing.B) {
	for _, test := range stringMatchTests {
		insp, err := newStringMatch(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkStringMatchByte(b, insp, message)
			},
		)
	}
}

func FuzzTestStringMatch(f *testing.F) {
	testcases := [][]byte{
		[]byte("Test"),
		[]byte("-Test"),
		[]byte("AnotherTest"),
		[]byte("123Test"),
		[]byte(""),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newStringMatch(ctx, config.Config{
			Settings: map[string]interface{}{
				"pattern": "^Test",
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
