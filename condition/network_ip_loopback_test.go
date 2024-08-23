package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Inspector = &networkIPLoopback{}

var networkIPLoopbackTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{},
		[]byte("127.0.0.1"),
		true,
	},
	{
		"fail",
		config.Config{},
		[]byte("8.8.8.8"),
		false,
	},
}

func TestNetworkIPLoopback(t *testing.T) {
	ctx := context.TODO()

	for _, test := range networkIPLoopbackTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNetworkIPLoopback(ctx, test.cfg)
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

func benchmarkNetworkIPLoopbackByte(b *testing.B, insp *networkIPLoopback, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkNetworkIPLoopbackByte(b *testing.B) {
	for _, test := range networkIPLoopbackTests {
		insp, err := newNetworkIPLoopback(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNetworkIPLoopbackByte(b, insp, message)
			},
		)
	}
}
