package process

import (
	"bytes"
	"context"
	"testing"
)

func TestHash(t *testing.T) {
	var tests = []struct {
		proc     Hash
		test     []byte
		expected []byte
	}{
		{
			Hash{
				Input: Input{
					Key: "hash",
				},
				Options: HashOptions{
					Algorithm: "sha256",
				},
				Output: Output{
					Key: "hash",
				},
			},
			[]byte(`{"hash":"hello"}`),
			[]byte(`{"hash":"2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"}`),
		},
		{
			Hash{
				Input: Input{
					Key: "@this",
				},
				Options: HashOptions{
					Algorithm: "sha256",
				},
				Output: Output{
					Key: "hash",
				},
			},
			[]byte(`{"hash":"hello"}`),
			[]byte(`{"hash":"a3db7a8b21d21028ed03f0dcbff12d279c06112c8294b32590bf6033d734eb2a"}`),
		},
		// array support
		{
			Hash{
				Input: Input{
					Key: "hash",
				},
				Options: HashOptions{
					Algorithm: "sha256",
				},
				Output: Output{
					Key: "hash",
				},
			},
			[]byte(`{"hash":["hello","goodbye"]}`),
			[]byte(`{"hash":["2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824","82e35a63ceba37e9646434c5dd412ea577147f1e4a41ccde1614253187e3dbf9"]}`),
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
