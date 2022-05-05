package condition

import (
	"testing"
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
	for _, testing := range ipTests {
		check, _ := testing.inspector.Inspect(testing.test)
		if testing.expected != check {
			t.Logf("expected %v, got %v, %v", testing.expected, check, string(testing.test))
			t.Fail()
		}
	}
}

func benchmarkIPByte(b *testing.B, inspector IP, test []byte) {
	for i := 0; i < b.N; i++ {
		inspector.Inspect(test)
	}
}

func BenchmarkIPByte(b *testing.B) {
	for _, test := range ipTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkIPByte(b, test.inspector, test.test)
			},
		)
	}
}
