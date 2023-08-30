package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modTime{}

var modTimeTestSetFmt = "2006-01-02T15:04:05.000000Z"

var modTimeTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"data pattern",
		config.Config{
			Settings: map[string]interface{}{
				"format":     "2006-01-02T15:04:05Z",
				"set_format": modTimeTestSetFmt,
			},
		},
		[]byte(`2021-03-06T00:02:57Z`),
		[][]byte{
			[]byte(`2021-03-06T00:02:57.000000Z`),
		},
		nil,
	},
	{
		"data unix",
		config.Config{
			Settings: map[string]interface{}{
				"format":     "unix",
				"set_format": modTimeTestSetFmt,
			},
		},
		[]byte(`1639877490.061`),
		[][]byte{
			[]byte(`2021-12-19T01:31:30.061000Z`),
		},
		nil,
	},
	{
		"data unix_milli",
		config.Config{
			Settings: map[string]interface{}{
				"format":     "unix",
				"set_format": "unix_milli",
			},
		},
		[]byte(`1639877490.061`),
		[][]byte{
			[]byte(`1639877490061`),
		},
		nil,
	},
	{
		"object pattern",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "time",
					"set_key": "time",
				},
				"format":     "2006-01-02T15:04:05Z",
				"set_format": modTimeTestSetFmt,
			},
		},
		[]byte(`{"time":"2021-03-06T00:02:57Z"}`),
		[][]byte{
			[]byte(`{"time":"2021-03-06T00:02:57.000000Z"}`),
		},
		nil,
	},
	{
		"object unix",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "time",
					"set_key": "time",
				},
				"format":     modTimeTestSetFmt,
				"set_format": "unix",
			},
		},
		[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
		[][]byte{
			[]byte(`{"time":1639877490}`),
		},
		nil,
	},
	{
		"object unix_milli",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "time",
					"set_key": "time",
				},
				"format":     modTimeTestSetFmt,
				"set_format": "unix_milli",
			},
		},
		[]byte(`{"time":"2022-06-05T20:07:12.263000Z"}`),
		[][]byte{
			[]byte(`{"time":1654459632263}`),
		},
		nil,
	},
	{
		"object from unix",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "time",
					"set_key": "time",
				},
				"format":     "unix",
				"set_format": modTimeTestSetFmt,
			},
		},
		[]byte(`{"time":1639877490}`),
		[][]byte{
			[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
		},
		nil,
	},
	{
		"object from unix_milli",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "time",
					"set_key": "time",
				},
				"format":     "unix_milli",
				"set_format": modTimeTestSetFmt,
			},
		},
		[]byte(`{"time":1654459632263}`),
		[][]byte{
			[]byte(`{"time":"2022-06-05T20:07:12.263000Z"}`),
		},
		nil,
	},
	{
		"object unix to unix_milli",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "time",
					"set_key": "time",
				},
				"format":     "unix",
				"set_format": "unix_milli",
			},
		},
		[]byte(`{"time":1639877490}`),
		[][]byte{
			[]byte(`{"time":1639877490000}`),
		},
		nil,
	},
	{
		"object offset conversion",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "time",
					"set_key": "time",
				},
				"format":     "2006-Jan-02 Monday 03:04:05 -0700",
				"set_format": "2006-Jan-02 Monday 03:04:05 -0700",
			},
		},
		[]byte(`{"time":"2020-Jan-29 Wednesday 12:19:25 -0500"}`),
		[][]byte{
			[]byte(`{"time":"2020-Jan-29 Wednesday 05:19:25 +0000"}`),
		},
		nil,
	},
	{
		"object offset to local conversion",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "time",
					"set_key": "time",
				},
				"format":       "2006-Jan-02 Monday 03:04:05 -0700",
				"set_format":   "2006-Jan-02 Monday 03:04:05 PM",
				"set_location": "America/New_York",
			},
		},
		// 12:19:25 AM in Pacific Standard time
		[]byte(`{"time":"2020-Jan-29 Wednesday 00:19:25 -0800"}`),
		// 03:19:25 AM in Eastern Standard time
		[][]byte{
			[]byte(`{"time":"2020-Jan-29 Wednesday 03:19:25 AM"}`),
		},
		nil,
	},
	{
		"object local to local conversion",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "time",
					"set_key": "time",
				},
				"format":       "2006-Jan-02 Monday 03:04:05",
				"location":     "America/Los_Angeles",
				"set_format":   "2006-Jan-02 Monday 03:04:05",
				"set_location": "America/New_York",
			},
		},
		// 12:19:25 AM in Pacific Standard time
		[]byte(`{"time":"2020-Jan-29 Wednesday 00:19:25"}`),
		// 03:19:25 AM in Eastern Standard time
		[][]byte{
			[]byte(`{"time":"2020-Jan-29 Wednesday 03:19:25"}`),
		},
		nil,
	},
}

func TestModTime(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modTimeTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModTime(ctx, test.cfg)
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

func benchmarkModTime(b *testing.B, tf *modTime, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModTime(b *testing.B) {
	for _, test := range modTimeTests {
		tf, err := newModTime(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModTime(b, tf, test.test)
			},
		)
	}
}
