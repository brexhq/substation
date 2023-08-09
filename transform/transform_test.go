package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var transformTests = []struct {
	name     string
	conf     config.Config
	test     []byte
	expected [][]byte
}{
	{
		"copy",
		config.Config{
			Type: "proc_copy",
			Settings: map[string]interface{}{
				"set_key": "foo",
			},
		},
		[]byte(`bar`),
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
	},
	{
		"insert",
		config.Config{
			Type: "proc_insert",
			Settings: map[string]interface{}{
				"set_key": "foo",
				"value":   "bar",
			},
		},
		[]byte(`{"hello":"world"}`),
		[][]byte{
			[]byte(`{"hello":"world","foo":"bar"}`),
		},
	},
	{
		"gzip",
		config.Config{
			Type: "proc_gzip",
			Settings: map[string]interface{}{
				"direction": "from",
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 170, 86, 202, 72, 205, 201, 201, 87, 178, 82, 74, 207, 207, 79, 73, 170, 76, 85, 170, 5, 4, 0, 0, 255, 255, 214, 182, 196, 150, 19, 0, 0, 0},
		[][]byte{
			[]byte(`{"hello":"goodbye"}`),
		},
	},
	{
		"base64",
		config.Config{
			Type: "proc_base64",
			Settings: map[string]interface{}{
				"direction": "from",
			},
		},
		[]byte(`eyJoZWxsbyI6IndvcmxkIn0=`),
		[][]byte{
			[]byte(`{"hello":"world"}`),
		},
	},
	{
		"split",
		config.Config{
			Type: "proc_split",
			Settings: map[string]interface{}{
				"key":       "foo",
				"set_key":   "foo",
				"separator": ".",
			},
		},
		[]byte(`{"foo":"bar.baz"}`),
		[][]byte{
			[]byte(`{"foo":["bar","baz"]}`),
		},
	},
	{
		"pretty_print",
		config.Config{
			Type: "proc_pretty_print",
			Settings: map[string]interface{}{
				"direction": "to",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[][]byte{
			[]byte(`{
  "foo": "bar"
}
`),
		},
	},
	{
		"time",
		config.Config{
			Type: "proc_time",
			Settings: map[string]interface{}{
				"key":        "foo",
				"set_key":    "foo",
				"format":     "unix",
				"set_format": "2006-01-02T15:04:05.000000Z",
			},
		},
		[]byte(`{"foo":1639877490}`),
		[][]byte{
			[]byte(`{"foo":"2021-12-19T01:31:30.000000Z"}`),
		},
	},
}

func TestTransform(t *testing.T) {
	ctx := context.TODO()
	for _, test := range transformTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			tform, err := New(ctx, test.conf)
			if err != nil {
				t.Error(err)
			}

			result, err := tform.Transform(ctx, message)
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
