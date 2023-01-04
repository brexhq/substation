package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var ipTests = []struct {
	name      string
	inspector inspIP
	test      []byte
	expected  bool
}{
	{
		"json",
		inspIP{
			condition: condition{
				Key: "ip_address",
			},
			Options: inspIPOptions{
				Type: "private",
			},
		},
		[]byte(`{"ip_address":"192.168.1.2"}`),
		true,
	},
	{
		"valid",
		inspIP{
			Options: inspIPOptions{
				Type: "valid",
			},
		},
		[]byte("192.168.1.2"),
		true,
	},
	{
		"invalid",
		inspIP{
			Options: inspIPOptions{
				Type: "valid",
			},
		},
		[]byte("foo"),
		false,
	},
	{
		"multicast",
		inspIP{
			Options: inspIPOptions{
				Type: "multicast",
			},
		},
		[]byte("224.0.0.12"),
		true,
	},
	{
		"multicast_link_local",
		inspIP{
			Options: inspIPOptions{
				Type: "multicast_link_local",
			},
		},
		[]byte("224.0.0.12"),
		true,
	},
	{
		"unicast_global",
		inspIP{
			Options: inspIPOptions{
				Type: "unicast_global",
			},
		},
		[]byte("8.8.8.8"),
		true,
	},
	{
		"private",
		inspIP{
			Options: inspIPOptions{
				Type: "private",
			},
		},
		[]byte("8.8.8.8"),
		false,
	},
	{
		"unicast_link_local",
		inspIP{
			Options: inspIPOptions{
				Type: "unicast_link_local",
			},
		},
		[]byte("169.254.255.255"),
		true,
	},
	{
		"loopback",
		inspIP{
			Options: inspIPOptions{
				Type: "loopback",
			},
		},
		[]byte("127.0.0.1"),
		true,
	},
	{
		"unspecified",
		inspIP{
			Options: inspIPOptions{
				Type: "unspecified",
			},
		},
		[]byte("0.0.0.0"),
		true,
	},
}

func TestIP(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range ipTests {
		var _ Inspector = test.inspector

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

func benchmarkIPByte(b *testing.B, inspector inspIP, capsule config.Capsule) {
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
