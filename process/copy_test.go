package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
)

var copyTests = []struct {
	name     string
	proc     Copy
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON",
		Copy{
			InputKey:  "foo",
			OutputKey: "baz",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"bar","baz":"bar"}`),
		nil,
	},
	{
		"from JSON",
		Copy{
			InputKey: "foo",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`bar`),
		nil,
	},
	{
		"from JSON nested",
		Copy{
			InputKey: "foo",
		},
		[]byte(`{"foo":{"bar":"baz"}}`),
		[]byte(`{"bar":"baz"}`),
		nil,
	},
	{
		"to JSON utf8",
		Copy{
			OutputKey: "bar",
		},
		[]byte(`baz`),
		[]byte(`{"bar":"baz"}`),
		nil,
	},
	{
		"to JSON zlib",
		Copy{
			OutputKey: "bar",
		},
		[]byte{120, 156, 5, 192, 49, 13, 0, 0, 0, 194, 48, 173, 76, 2, 254, 143, 166, 29, 2, 93, 1, 54},
		[]byte(`{"bar":"eJwFwDENAAAAwjCtTAL+j6YdAl0BNg=="}`),
		nil,
	},
}

func TestCopy(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range copyTests {
		cap.SetData(test.test)

		res, err := test.proc.Apply(ctx, cap)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res.GetData(), test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res.GetData())
			t.Fail()
		}
	}
}

func benchmarkCopy(b *testing.B, applicator Copy, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkCopy(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range copyTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkCopy(b, test.proc, cap)
			},
		)
	}
}
