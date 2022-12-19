package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var ipTests = []struct {
	name      string
	inspector ip
	test      []byte
	expected  bool
}{
	{
		"json",
		ip{
			condition: condition{
				Key: "ip_address",
			},
			Options: ipOptions{
				Type: "private",
			},
		},
		[]byte(`{"ip_address":"192.168.1.2"}`),
		true,
	},
	{
		"valid",
		ip{
			Options: ipOptions{
				Type: "valid",
			},
		},
		[]byte("192.168.1.2"),
		true,
	},
	{
		"invalid",
		ip{
			Options: ipOptions{
				Type: "valid",
			},
		},
		[]byte("foo"),
		false,
	},
	{
		"multicast",
		ip{
			Options: ipOptions{
				Type: "multicast",
			},
		},
		[]byte("224.0.0.12"),
		true,
	},
	{
		"multicast_link_local",
		ip{
			Options: ipOptions{
				Type: "multicast_link_local",
			},
		},
		[]byte("224.0.0.12"),
		true,
	},
	{
		"unicast_global",
		ip{
			Options: ipOptions{
				Type: "unicast_global",
			},
		},
		[]byte("8.8.8.8"),
		true,
	},
	{
		"private",
		ip{
			Options: ipOptions{
				Type: "private",
			},
		},
		[]byte("8.8.8.8"),
		false,
	},
	{
		"unicast_link_local",
		ip{
			Options: ipOptions{
				Type: "unicast_link_local",
			},
		},
		[]byte("169.254.255.255"),
		true,
	},
	{
		"loopback",
		ip{
			Options: ipOptions{
				Type: "loopback",
			},
		},
		[]byte("127.0.0.1"),
		true,
	},
	{
		"unspecified",
		ip{
			Options: ipOptions{
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

func benchmarkIPByte(b *testing.B, inspector ip, capsule config.Capsule) {
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
