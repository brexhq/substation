package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &networkIPValid{}

var networkIPValidTests = []struct {
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
}

func TestNetworkIPValid(t *testing.T) {
	ctx := context.TODO()

	for _, test := range networkIPValidTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNetworkIPValid(ctx, test.cfg)
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

func benchmarkNetworkIPValidByte(b *testing.B, insp *networkIPValid, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkNetworkIPValidByte(b *testing.B) {
	for _, test := range networkIPValidTests {
		insp, err := newNetworkIPValid(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNetworkIPValidByte(b, insp, message)
			},
		)
	}
}

func FuzzTestNetworkIPValid(f *testing.F) {
	testcases := [][]byte{
		[]byte("192.168.1.1"),
		[]byte("8.8.8.8"),
		[]byte("255.255.255.255"),
		[]byte("::1"),
		[]byte("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newNetworkIPValid(ctx, config.Config{})
		if err != nil {
			return
		}

		_, err = insp.Condition(ctx, message)
		if err != nil {
			return
		}
	})
}
