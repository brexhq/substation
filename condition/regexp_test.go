package condition

import (
	"testing"
)

func TestRegExp(t *testing.T) {
	var tests = []struct {
		con      RegExp
		test     []byte
		expected bool
	}{
		{
			RegExp{
				Expression: "^Test",
			},
			[]byte("Test"),
			true,
		},
		{
			RegExp{
				Expression: "^Test",
			},
			[]byte("-Test"),
			false,
		},
		{
			RegExp{
				Negate:     true,
				Expression: "XYZ",
			},
			[]byte("ABC"),
			true,
		},
		{
			RegExp{
				Negate:     true,
				Expression: "ABC",
			},
			[]byte("ABC"),
			false,
		},
	}

	for _, rt := range tests {
		check, _ := rt.con.Inspect(rt.test)

		if rt.expected != check {
			t.Logf("expected %v, got %v", rt.expected, check)
			t.Fail()
		}
	}
}
