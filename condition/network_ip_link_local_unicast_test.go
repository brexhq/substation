package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &networkIPLinkLocalUnicast{}

var networkIPLinkLocalUnicastTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{},
		[]byte("169.254.255.255"),
		true,
	},
}

func TestNetworkIPLinkLocalUnicast(t *testing.T) {
	ctx := context.TODO()

	for _, test := range networkIPLinkLocalUnicastTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNetworkIPLinkLocalUnicast(ctx, test.cfg)
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

func benchmarkNetworkIPLinkLocalUnicastByte(b *testing.B, insp *networkIPLinkLocalUnicast, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkNetworkIPLinkLocalUnicastByte(b *testing.B) {
	for _, test := range networkIPLinkLocalUnicastTests {
		insp, err := newNetworkIPLinkLocalUnicast(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNetworkIPLinkLocalUnicastByte(b, insp, message)
			},
		)
	}
}

func FuzzTestNetworkIPLinkLocalUnicast(f *testing.F) {
	testcases := [][]byte{
		[]byte("169.254.255.255"),
		[]byte("169.254.0.1"),
		[]byte("192.168.1.1"),
		[]byte("10.0.0.1"),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newNetworkIPLinkLocalUnicast(ctx, config.Config{})
		if err != nil {
			return
		}

		_, err = insp.Condition(ctx, message)
		if err != nil {
			return
		}
	})
}
