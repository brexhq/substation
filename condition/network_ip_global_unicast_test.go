package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &networkIPGlobalUnicast{}

var networkIPGlobalUnicastTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{},
		[]byte("8.8.8.8"),
		true,
	},
}

func TestNetworkIPGlobalUnicast(t *testing.T) {
	ctx := context.TODO()

	for _, test := range networkIPGlobalUnicastTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNetworkIPGlobalUnicast(ctx, test.cfg)
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

func benchmarkNetworkIPGlobalUnicastByte(b *testing.B, insp *networkIPGlobalUnicast, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkNetworkIPGlobalUnicastByte(b *testing.B) {
	for _, test := range networkIPGlobalUnicastTests {
		insp, err := newNetworkIPGlobalUnicast(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNetworkIPGlobalUnicastByte(b, insp, message)
			},
		)
	}
}
