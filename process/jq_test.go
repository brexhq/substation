package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var jsonTests = []struct {
	name     string
	proc     _jq
	test     []byte
	expected []byte
	err      error
}{
	{
		"recursively removed null and empty values",
		_jq{
			process: process{},
			Options: _jqOptions{
				Query: `walk( if type == "object" then 
				with_entries( select( 
					(.value != "") and 
					(.value != {}) and
					(.value != null)
				) ) 
			else 
				. end)`,
			},
		},
		[]byte(`{"foo":{"bar":{"baz":""}},"qux":null,"quux":"corge"}`),
		[]byte(`{"quux":"corge"}`),
		nil,
	},
}

func TestJq(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range jsonTests {
		capsule.SetData(test.test)

		result, err := test.proc.Apply(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkJq(b *testing.B, applier _jq, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkJq(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range jsonTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkJq(b, test.proc, capsule)
			},
		)
	}
}
