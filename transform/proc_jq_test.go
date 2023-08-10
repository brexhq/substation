package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procJQ{}

var procJQTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"access",
		config.Config{
			Settings: map[string]interface{}{
				"query": `.a`,
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`"b"`),
		},
		nil,
	},
	{
		"access",
		config.Config{
			Settings: map[string]interface{}{
				"query": `.a, .c`,
			},
		},
		[]byte(`{"a":"b","c":"d"}`),
		[][]byte{
			[]byte(`["b","d"]`),
		},
		nil,
	},
	{
		"access",
		config.Config{
			Settings: map[string]interface{}{
				"query": `.a`,
			},
		},
		[]byte(`{"a":{"b":"c"}}`),
		[][]byte{
			[]byte(`{"b":"c"}`),
		},
		nil,
	},
	{
		"array",
		config.Config{
			Settings: map[string]interface{}{
				"query": `.a`,
			},
		},
		[]byte(`{"a":["b","c","d"]}`),
		[][]byte{
			[]byte(`["b","c","d"]`),
		},
		nil,
	},
	{
		"slice",
		config.Config{
			Settings: map[string]interface{}{
				"query": `.a[-1:]`,
			},
		},
		[]byte(`{"a":["b","c","d","e","f","g"]}`),
		[][]byte{
			[]byte(`["g"]`),
		},
		nil,
	},
	{
		"recursion",
		config.Config{
			Settings: map[string]interface{}{
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
		[]byte(`{"a":{"b":{"c":""}},"d":null,"e":"f"}`),
		[][]byte{
			[]byte(`{"e":"f"}`),
		},
		nil,
	},
}

func TestProcJQ(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procJQTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcJQ(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Transform(ctx, message)
			if err != nil {
				t.Error(err)
			}

			var data [][]byte
			for _, c := range result {
				data = append(data, c.Data())
			}

			if !reflect.DeepEqual(data, test.expected) {
				t.Errorf("expected %s, got %s", test.expected, data)
			}
		})
	}
}

func benchmarkProcJQ(b *testing.B, tform *procJQ, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcJQ(b *testing.B) {
	for _, test := range procJQTests {
		proc, err := newProcJQ(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcJQ(b, proc, test.test)
			},
		)
	}
}
