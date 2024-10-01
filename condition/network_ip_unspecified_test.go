package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &networkIPUnspecified{}

var networkIPUnspecifiedTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{},
		[]byte("0.0.0.0"),
		true,
	},
}

func TestNetworkIPUnspecified(t *testing.T) {
	ctx := context.TODO()

	for _, test := range networkIPUnspecifiedTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNetworkIPUnspecified(ctx, test.cfg)
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

func benchmarkNetworkIPUnspecifiedByte(b *testing.B, insp *networkIPUnspecified, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkNetworkIPUnspecifiedByte(b *testing.B) {
	for _, test := range networkIPUnspecifiedTests {
		insp, err := newNetworkIPUnspecified(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNetworkIPUnspecifiedByte(b, insp, message)
			},
		)
	}
}

func FuzzTestNetworkIPUnspecified(f *testing.F) {
	testcases := [][]byte{
		[]byte("0.0.0.0"),
		[]byte("192.168.1.1"),
		[]byte("::"),
		[]byte("255.255.255.255"),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newNetworkIPUnspecified(ctx, config.Config{})
		if err != nil {
			return
		}

		_, err = insp.Condition(ctx, message)
		if err != nil {
			return
		}
	})
}
