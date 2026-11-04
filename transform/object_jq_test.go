package transform

import (
	"context"
	"reflect"
	"strings"
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
	{
		"large integer",
		config.Config{
			Settings: map[string]interface{}{
				"filter": `.`,
			},
		},
		[]byte(`{"id":1234567890123456789}`),
		[][]byte{
			[]byte(`{"id":1234567890123456789}`),
		},
		nil,
	},
	{
		"integer above 2^64",
		config.Config{
			Settings: map[string]interface{}{
				"filter": `.`,
			},
		},
		[]byte(`{"id":123456789012345678901234567890}`),
		[][]byte{
			[]byte(`{"id":123456789012345678901234567890}`),
		},
		nil,
	},
	{
		"negative large integer",
		config.Config{
			Settings: map[string]interface{}{
				"filter": `.`,
			},
		},
		[]byte(`{"id":-123456789012345678901234567890}`),
		[][]byte{
			[]byte(`{"id":-123456789012345678901234567890}`),
		},
		nil,
	},
	{
		"large integer arithmetic",
		config.Config{
			Settings: map[string]interface{}{
				"filter": `.id + 1`,
			},
		},
		[]byte(`{"id":123456789012345678901234567890}`),
		[][]byte{
			[]byte(`123456789012345678901234567891`),
		},
		nil,
	},
	{
		"float",
		config.Config{
			Settings: map[string]interface{}{
				"filter": `.`,
			},
		},
		[]byte(`{"x":1.5}`),
		[][]byte{
			[]byte(`{"x":1.5}`),
		},
		nil,
	},
	{
		"exponent",
		config.Config{
			Settings: map[string]interface{}{
				"filter": `.x * 2`,
			},
		},
		[]byte(`{"x":1e3}`),
		[][]byte{
			[]byte(`2000`),
		},
		nil,
	},
	{
		"negative zero",
		config.Config{
			Settings: map[string]interface{}{
				"filter": `.`,
			},
		},
		[]byte(`{"x":-0}`),
		[][]byte{
			[]byte(`{"x":-0}`),
		},
		nil,
	},
	{
		"negative zero tostring",
		config.Config{
			Settings: map[string]interface{}{
				"filter": `.x | tostring`,
			},
		},
		[]byte(`{"x":-0}`),
		[][]byte{
			[]byte(`"-0"`),
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

func TestObjectJQInvalidInput(t *testing.T) {
	ctx := context.TODO()
	tf, err := newObjectJQ(ctx, config.Config{
		Settings: map[string]interface{}{
			"filter": `.`,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// An empty want accepts any error.
	for _, test := range []struct {
		data string
		want string
	}{
		{``, ""},
		{`{"a":`, ""},
		{`not json`, ""},
		{`{"x":1e400}`, "value out of range"},
		{`{"x":1e400,"x":1}`, "value out of range"},
		{`[1,{"x":[-1.5e400]}]`, "value out of range"},
		{`{"a":1} x`, "invalid character 'x'"},
		{`{"a":1} {"b":2}`, "unexpected data after top-level value"},
		{`{"a":1} {"b":`, "unexpected data after top-level value"},
		{`{} "extra-value"`, "unexpected data after top-level value"},
	} {
		msg := message.New().SetData([]byte(test.data))
		_, err := tf.Transform(ctx, msg)
		if err == nil {
			t.Errorf("input %q: expected error, got nil", test.data)
			continue
		}
		if !strings.Contains(err.Error(), test.want) {
			t.Errorf("input %q: expected error containing %q, got %v", test.data, test.want, err)
		}
	}

	// The trailing value must not appear in the error, since errors are logged.
	if _, err := tf.Transform(ctx, message.New().SetData([]byte(`{} "extra-value"`))); err == nil || strings.Contains(err.Error(), "extra-value") {
		t.Errorf("trailing string: expected error without the value, got %v", err)
	}

	// A float overflow fails as json.Unmarshal did, rather than becoming +Inf.
	tostring, err := newObjectJQ(ctx, config.Config{
		Settings: map[string]interface{}{
			"filter": `.x | tostring`,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tostring.Transform(ctx, message.New().SetData([]byte(`{"x":1e400}`))); err == nil {
		t.Errorf("1e400 | tostring: expected error, got nil")
	}

	// Trailing whitespace is allowed, as with json.Unmarshal.
	msg := message.New().SetData([]byte("{\"a\":1}\n"))
	if _, err := tf.Transform(ctx, msg); err != nil {
		t.Errorf("trailing whitespace: unexpected error: %v", err)
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

func FuzzTestObjectJQ(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":"b"}`),
		[]byte(`{"a":{"b":"c"}}`),
		[]byte(`{"array":[1,2,3]}`),
		[]byte(`{"a":"{\"b\":\"c\"}"}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Use a sample configuration for the transformer
		tf, err := newObjectJQ(ctx, config.Config{
			Settings: map[string]interface{}{
				"filter": `.a`,
			},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}
	})
}
