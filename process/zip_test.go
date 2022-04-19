package process

import (
	"bytes"
	"context"
	"testing"
)

func TestZip(t *testing.T) {
	var tests = []struct {
		proc Zip
		test []byte
		// the order of the zipped output is inconsistent, so we check for a match anywhere in this slice
		expected [][]byte
	}{
		{
			Zip{
				Input: Inputs{
					Keys: []string{"names", "sizes"},
				},
				Output: Output{
					Key: "zipped",
				},
			},
			[]byte(`{"names":["foo","bar"],"sizes":[123,456]}`),
			[][]byte{
				[]byte(`{"names":["foo","bar"],"sizes":[123,456],"zipped":[["foo",123],["bar",456]]}`),
				[]byte(`{"names":["foo","bar"],"sizes":[123,456],"zipped":[["bar",456],["foo",123]]}`),
			},
		},
		{
			Zip{
				Input: Inputs{
					Keys: []string{"names", "sizes"},
				},
				Options: ZipOptions{
					Keys: []string{"name.test", "size"},
				},
				Output: Output{
					Key: "zipped",
				},
			},
			[]byte(`{"names":["foo","bar"],"sizes":[123,456]}`),
			[][]byte{
				[]byte(`{"names":["foo","bar"],"sizes":[123,456],"zipped":[{"name":{"test":"foo"},"size":123},{"name":{"test":"bar"},"size":456}]}`),
				[]byte(`{"names":["foo","bar"],"sizes":[123,456],"zipped":[{"name":{"test":"bar"},"size":456},{"name":{"test":"foo"},"size":123}]}`),
			},
		},
	}

	for _, test := range tests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil {
			t.Logf("%v", err)
			t.Fail()
		}

		pass := false
		for _, x := range test.expected {
			if c := bytes.Compare(res, x); c == 0 {
				pass = true
			}
		}

		if !pass {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}
