package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ inspector = &networkIPLinkLocalMulticast{}

var networkIPLinkLocalMulticastTests = []struct {
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

func TestNetworkIPLinkLocalMulticast(t *testing.T) {
	ctx := context.TODO()

	for _, test := range networkIPLinkLocalMulticastTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNetworkIPLinkLocalMulticast(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.test))
			}
		})
	}
}

func benchmarkNetworkIPLinkLocalMulticastByte(b *testing.B, insp *networkIPLinkLocalMulticast, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkNetworkIPLinkLocalMulticastByte(b *testing.B) {
	for _, test := range networkIPLinkLocalMulticastTests {
		insp, err := newNetworkIPLinkLocalMulticast(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNetworkIPLinkLocalMulticastByte(b, insp, message)
			},
		)
	}
}
