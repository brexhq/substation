package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &networkIPMulticast{}

var networkIPMulticastTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{},
		[]byte("224.0.0.12"),
		true,
	},
}

func TestNetworkIPMulticast(t *testing.T) {
	ctx := context.TODO()

	for _, test := range networkIPMulticastTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNetworkIPMulticast(ctx, test.cfg)
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

func benchmarkNetworkIPMulticastByte(b *testing.B, insp *networkIPMulticast, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkNetworkIPMulticastByte(b *testing.B) {
	for _, test := range networkIPMulticastTests {
		insp, err := newNetworkIPMulticast(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNetworkIPMulticastByte(b, insp, message)
			},
		)
	}
}

func FuzzTestNetworkIPMulticast(f *testing.F) {
	testcases := [][]byte{
		[]byte("224.0.0.12"),
		[]byte("239.255.255.255"),
		[]byte("192.168.1.1"),
		[]byte("10.0.0.1"),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newNetworkIPMulticast(ctx, config.Config{})
		if err != nil {
			return
		}

		_, err = insp.Condition(ctx, message)
		if err != nil {
			return
		}
	})
}
