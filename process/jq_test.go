package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var jsonTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"access",
		config.Config{
			Type: "jq",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"query": `.a`,
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[]byte(`"b"`),
		nil,
	},
	{
		"access",
		config.Config{
			Type: "jq",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"query": `.a, .c`,
				},
			},
		},
		[]byte(`{"a":"b","c":"d"}`),
		[]byte(`["b","d"]`),
		nil,
	},
	{
		"access",
		config.Config{
			Type: "jq",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"query": `.a`,
				},
			},
		},
		[]byte(`{"a":{"b":"c"}}`),
		[]byte(`{"b":"c"}`),
		nil,
	},
	{
		"array",
		config.Config{
			Type: "jq",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"query": `.a`,
				},
			},
		},
		[]byte(`{"a":["b","c","d"]}`),
		[]byte(`["b","c","d"]`),
		nil,
	},
	{
		"slice",
		config.Config{
			Type: "jq",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"query": `.a[-1:]`,
				},
			},
		},
		[]byte(`{"a":["b","c","d","e","f","g"]}`),
		[]byte(`["g"]`),
		nil,
	},
	{
		"recursion",
		config.Config{
			Type: "jq",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"query": `walk( if type == "object" then 
					with_entries( select( 
						(.value != "") and 
						(.value != {}) and
						(.value != null)
					) ) 
				else 
					. end)`,
				},
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
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcJQ(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Apply(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(result.Data(), test.expected) {
				t.Errorf("expected %s, got %s", test.expected, result.Data())
			}
		})
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
		proc, err := newProcJQ(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkJq(b, proc, capsule)
			},
		)
	}
}
