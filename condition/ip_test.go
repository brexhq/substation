package condition

import (
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
	cap := config.NewCapsule()
	for _, test := range ipTests {
		cap.SetData(test.test)
		check, _ := test.inspector.Inspect(cap)

		if test.expected != check {
			t.Logf("expected %v, got %v, %v", test.expected, check, string(test.test))
			t.Fail()
		}
	}
}

func benchmarkIPByte(b *testing.B, inspector IP, cap config.Capsule) {
	for i := 0; i < b.N; i++ {
		inspector.Inspect(cap)
	}
}

func BenchmarkIPByte(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range ipTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkIPByte(b, test.inspector, cap)
			},
		)
	}
}
