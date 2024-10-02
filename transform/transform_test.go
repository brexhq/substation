package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var transformTests = []struct {
	name     string
	conf     config.Config
	test     []byte
	expected [][]byte
}{
	{
		"object_copy",
		config.Config{
			Type: "object_copy",
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"target_key": "a",
				},
			},
		},
		[]byte(`b`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},
	{
		"object_insert",
		config.Config{
			Type: "object_insert",
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"target_key": "c",
				},
				"value": "d",
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"b","c":"d"}`),
		},
	},
	{
		"format_from_gzip",
		config.Config{
			Type: "format_from_gzip",
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 170, 86, 202, 72, 205, 201, 201, 87, 178, 82, 74, 207, 207, 79, 73, 170, 76, 85, 170, 5, 4, 0, 0, 255, 255, 214, 182, 196, 150, 19, 0, 0, 0},
		[][]byte{
			[]byte(`{"hello":"goodbye"}`),
		},
	},
	{
		"format_from_base64",
		config.Config{
			Type: "format_from_base64",
		},
		[]byte(`eyJoZWxsbyI6IndvcmxkIn0=`),
		[][]byte{
			[]byte(`{"hello":"world"}`),
		},
	},
	{
		"time_to_string",
		config.Config{
			Type: "time_to_string",
			Settings: map[string]interface{}{
				"format": "2006-01-02T15:04:05.000000Z",
			},
		},
		[]byte(`1639877490000000000`),
		[][]byte{
			[]byte(`2021-12-19T01:31:30.000000Z`),
		},
	},
}

func TestTransform(t *testing.T) {
	ctx := context.TODO()
	for _, test := range transformTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := New(ctx, test.conf)
			if err != nil {
				t.Error(err)
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

var truncateTTLTests = []struct {
	name     string
	test     []byte
	expected int64
}{
	{
		"unix millisecond",
		[]byte("1696482368492"),
		1696482368,
	},
	{
		"unix nanosecond",
		[]byte("1696482368492290"),
		1696482368,
	},
}

func TestTruncateTTL(t *testing.T) {
	for _, test := range truncateTTLTests {
		t.Run(test.name, func(t *testing.T) {
			tmp := bytesToValue(test.test)
			result := truncateTTL(tmp)

			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func FuzzTestTransform(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":"b"}`),
		[]byte(`{"a":{"b":"c"}}`),
		[]byte(`{"array":[1,2,3]}`),
		[]byte(`{"a":"{\"b\":\"c\"}"}`),
		[]byte(`eyJoZWxsbyI6IndvcmxkIn0=`),
		[]byte(`1639877490000000000`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		tf, err := New(ctx, config.Config{})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}
	})
}

func FuzzTestTransformApply(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":"b"}`),
		[]byte(`{"a":{"b":"c"}}`),
		[]byte(`{"array":[1,2,3]}`),
		[]byte(`{"a":"{\"b\":\"c\"}"}`),
		[]byte(`eyJoZWxsbyI6IndvcmxkIn0=`),
		[]byte(`1639877490000000000`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		tf, err := New(ctx, config.Config{})
		if err != nil {
			return
		}

		_, err = Apply(ctx, []Transformer{tf}, msg)
		if err != nil {
			return
		}
	})
}

func FuzzTestBytesToValue(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":"b"}`),
		[]byte(`{"a":{"b":"c"}}`),
		[]byte(`{"array":[1,2,3]}`),
		[]byte(`{"a":"{\"b\":\"c\"}"}`),
		[]byte(`eyJoZWxsbyI6IndvcmxkIn0=`),
		[]byte(`1639877490000000000`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		bytesToValue(data)
	})
}
