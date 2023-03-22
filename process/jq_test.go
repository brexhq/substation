package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var jsonTests = []struct {
	name     string
	proc     procJQ
	test     []byte
	expected []byte
	err      error
}{
	{
		"access",
		procJQ{
			process: process{},
			Options: procJQOptions{
				Query: `.a`,
			},
		},
		[]byte(`{"a":"b"}`),
		[]byte(`"b"`),
		nil,
	},
	{
		"access",
		procJQ{
			process: process{},
			Options: procJQOptions{
				Query: `.a, .c`,
			},
		},
		[]byte(`{"a":"b","c":"d"}`),
		[]byte(`["b","d"]`),
		nil,
	},
	{
		"access",
		procJQ{
			process: process{},
			Options: procJQOptions{
				Query: `.a`,
			},
		},
		[]byte(`{"a":{"b":"c"}}`),
		[]byte(`{"b":"c"}`),
		nil,
	},
	{
		"array",
		procJQ{
			process: process{},
			Options: procJQOptions{
				Query: `.a`,
			},
		},
		[]byte(`{"a":["b","c","d"]}`),
		[]byte(`["b","c","d"]`),
		nil,
	},
	{
		"slice",
		procJQ{
			process: process{},
			Options: procJQOptions{
				Query: `.a[-1:]`,
			},
		},
		[]byte(`{"a":["b","c","d","e","f","g"]}`),
		[]byte(`["g"]`),
		nil,
	},
	{
		"recursion",
		procJQ{
			process: process{},
			Options: procJQOptions{
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
		[]byte(`{"a":{"b":{"c":""}},"d":null,"e":"f"}`),
		[]byte(`{"e":"f"}`),
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

func benchmarkJq(b *testing.B, applier procJQ, test config.Capsule) {
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
