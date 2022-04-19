package condition

import (
	"testing"
)

func TestIP(t *testing.T) {
	var tests = []struct {
		inspector IP
		test      []byte
		expected  bool
	}{
		{
			IP{
				Function: "private",
				Key:      "foo",
			},
			[]byte(`{"foo":"192.168.1.2"}`),
			true,
		},
		{
			IP{
				Function: "multicast",
			},
			[]byte("224.0.0.12"),
			true,
		},
		{
			IP{
				Function: "multicast_link_local",
			},
			[]byte("224.0.0.12"),
			true,
		},
		{
			IP{
				Function: "unicast_global",
			},
			[]byte("8.8.8.8"),
			true,
		},
		{
			IP{
				Function: "private",
			},
			[]byte("8.8.8.8"),
			false,
		},
		{
			IP{
				Function: "unicast_link_local",
			},
			[]byte("169.254.255.255"),
			true,
		},
		{
			IP{
				Function: "loopback",
			},
			[]byte("127.0.0.1"),
			true,
		},
		{
			IP{
				Function: "unspecified",
			},
			[]byte("0.0.0.0"),
			true,
		},
	}

	for _, testing := range tests {
		check, _ := testing.inspector.Inspect(testing.test)

		if testing.expected != check {
			t.Logf("expected %v, got %v, %v", testing.expected, check, string(testing.test))
			t.Fail()
		}
	}
}
