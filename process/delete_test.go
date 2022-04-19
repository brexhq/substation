package process

import (
	"bytes"
	"context"
	"testing"
)

func TestDelete(t *testing.T) {
	var tests = []struct {
		proc     Delete
		test     []byte
		expected []byte
	}{
		// strings
		{
			Delete{
				Input: Input{
					Key: "delete",
				},
			},
			[]byte(`{"hello":"123","delete":"456"}`),
			[]byte(`{"hello":"123"}`),
		},
	}

	for _, test := range tests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil {
			t.Logf("%v", err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}
