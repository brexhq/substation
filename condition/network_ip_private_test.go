package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &networkIPPrivate{}

var networkIPPrivateTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{},
		[]byte("8.8.8.8"),
		false,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "ip_address",
				},
			},
		},
		[]byte(`{"ip_address":"192.168.1.2"}`),
		true,
	},
}

func TestNetworkIPPrivate(t *testing.T) {
	ctx := context.TODO()

	for _, test := range networkIPPrivateTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNetworkIPPrivate(ctx, test.cfg)
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

func benchmarkNetworkIPPrivateByte(b *testing.B, insp *networkIPPrivate, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkNetworkIPPrivateByte(b *testing.B) {
	for _, test := range networkIPPrivateTests {
		insp, err := newNetworkIPPrivate(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNetworkIPPrivateByte(b, insp, message)
			},
		)
	}
}

func FuzzTestNetworkIPPrivate(f *testing.F) {
	testcases := [][]byte{
		[]byte("8.8.8.8"),
		[]byte(`{"ip_address":"192.168.1.2"}`),
		[]byte("10.0.0.1"),
		[]byte("172.16.0.1"),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newNetworkIPPrivate(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "ip_address",
				},
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
