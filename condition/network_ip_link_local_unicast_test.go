package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

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

func benchmarkNetworkIPLinkLocalUnicastByte(b *testing.B, insp *networkIPLinkLocalUnicast, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
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
