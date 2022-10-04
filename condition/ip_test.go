package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var ipTests = []struct {
	name      string
	inspector IP
	test      []byte
	expected  bool
}{
	{
		"json",
		IP{
			Type: "private",
			Key:  "ip_address",
		},
		[]byte(`{"ip_address":"192.168.1.2"}`),
		true,
	},
	{
		"multicast",
		IP{
			Type: "multicast",
		},
		[]byte("224.0.0.12"),
		true,
	},
	{
		"multicast_link_local",
		IP{
			Type: "multicast_link_local",
		},
		[]byte("224.0.0.12"),
		true,
	},
	{
		"unicast_global",
		IP{
			Type: "unicast_global",
		},
		[]byte("8.8.8.8"),
		true,
	},
	{
		"private",
		IP{
			Type: "private",
		},
		[]byte("8.8.8.8"),
		false,
	},
	{
		"unicast_link_local",
		IP{
			Type: "unicast_link_local",
		},
		[]byte("169.254.255.255"),
		true,
	},
	{
		"loopback",
		IP{
			Type: "loopback",
		},
		[]byte("127.0.0.1"),
		true,
	},
	{
		"unspecified",
		IP{
			Type: "unspecified",
		},
		[]byte("0.0.0.0"),
		true,
	},
}

func TestIP(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range ipTests {
		capsule.SetData(test.test)

		check, err := test.inspector.Inspect(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if test.expected != check {
			t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.test))
		}
	}
}

func benchmarkIPByte(b *testing.B, inspector IP, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkIPByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range ipTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkIPByte(b, test.inspector, capsule)
			},
		)
	}
}
