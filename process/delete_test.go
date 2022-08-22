package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
)

var deleteTests = []struct {
	name     string
	proc     Delete
	test     []byte
	expected []byte
	err      error
}{
	{
		"string",
		Delete{
			InputKey: "baz",
		},
		[]byte(`{"foo":"bar","baz":"qux"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"JSON",
		Delete{
			InputKey: "baz",
		},
		[]byte(`{"foo":"bar","baz":{"qux":"quux"}}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
}

func TestDelete(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range convertTests {
		cap.SetData(test.test)

		res, err := test.proc.Apply(ctx, cap)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if !bytes.Equal(res.GetData(), test.expected) {
			t.Logf("expected %s, got %s", test.expected, res.GetData())
			t.Fail()
		}
	}
}

func benchmarkDelete(b *testing.B, applicator Delete, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkDelete(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range deleteTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkDelete(b, test.proc, cap)
			},
		)
	}
}
