package process

import (
	"bytes"
	"context"
	"testing"
)

func TestInsert(t *testing.T) {
	var tests = []struct {
		proc     Insert
		test     []byte
		expected []byte
	}{
		{
			Insert{
				Options: InsertOptions{
					Value: "goodbye",
				},
				Output: Output{
					Key: "hello",
				},
			},
			[]byte(``),
			[]byte(`{"hello":"goodbye"}`),
		},
		{
			Insert{
				Options: InsertOptions{
					Value: 10,
				},
				Output: Output{
					Key: "int",
				},
			},
			[]byte(`{"hello":"goodbye"}`),
			[]byte(`{"hello":"goodbye","int":10}`),
		},
		{
			Insert{
				Options: InsertOptions{
					Value: "goodbye",
				},
				Output: Output{
					Key: "__hidden",
				},
			},
			[]byte(`{"hello":"goodbye"}`),
			[]byte(`{"hello":"goodbye","__hidden":"goodbye"}`),
		},
		{
			Insert{
				Options: InsertOptions{
					Value: "goodbye",
				},
				Output: Output{
					Key: "__hidden.deeply.nested",
				},
			},
			[]byte(`{"hello":"goodbye"}`),
			[]byte(`{"hello":"goodbye","__hidden":{"deeply":{"nested":"goodbye"}}}`),
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
