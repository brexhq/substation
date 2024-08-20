package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &objectJQ{}

var objectJQTests = []struct {
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
				"filter": `.a`,
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
				"filter": `.a, .c`,
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
				"filter": `.a`,
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
				"filter": `.a`,
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
				"filter": `.a[-1:]`,
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
				"filter": `walk( if type == "object" then 
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

func TestObjectJQ(t *testing.T) {
	ctx := context.TODO()
	for _, test := range objectJQTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newObjectJQ(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			msg := message.New().SetData(test.test)
			result, err := tf.Transform(ctx, msg)
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

func benchmarkObjectJQ(b *testing.B, tf *objectJQ, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkObjectJQ(b *testing.B) {
	for _, test := range objectJQTests {
		tf, err := newObjectJQ(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkObjectJQ(b, tf, test.test)
			},
		)
	}
}
