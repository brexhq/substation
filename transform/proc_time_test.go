package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procTime{}

var setFmt = "2006-01-02T15:04:05.000000Z"

var procTimeTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"data string",
		config.Config{
			Type: "proc_time",
			Settings: map[string]interface{}{
				"format":     "2006-01-02T15:04:05Z",
				"set_format": setFmt,
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
			Type: "proc_time",
			Settings: map[string]interface{}{
				"format":     "unix",
				"set_format": setFmt,
			},
		},
		[]byte(`1639877490.061`),
		[][]byte{
			[]byte(`2021-12-19T01:31:30.061000Z`),
		},
		nil,
	},
	{
		"data unix to unix_milli",
		config.Config{
			Type: "proc_time",
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
		"JSON",
		config.Config{
			Type: "proc_time",
			Settings: map[string]interface{}{
				"key":        "time",
				"set_key":    "time",
				"format":     "2006-01-02T15:04:05Z",
				"set_format": setFmt,
			},
		},
		[]byte(`{"time":"2021-03-06T00:02:57Z"}`),
		[][]byte{
			[]byte(`{"time":"2021-03-06T00:02:57.000000Z"}`),
		},
		nil,
	},
	{
		"JSON from unix",
		config.Config{
			Type: "proc_time",
			Settings: map[string]interface{}{
				"key":        "time",
				"set_key":    "time",
				"format":     "unix",
				"set_format": setFmt,
			},
		},
		[]byte(`{"time":1639877490}`),
		[][]byte{
			[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
		},
		nil,
	},
	{
		"JSON to unix",
		config.Config{
			Type: "proc_time",
			Settings: map[string]interface{}{
				"key":        "time",
				"set_key":    "time",
				"format":     setFmt,
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
		"JSON from unix_milli",
		config.Config{
			Type: "proc_time",
			Settings: map[string]interface{}{
				"key":        "time",
				"set_key":    "time",
				"format":     "unix_milli",
				"set_format": setFmt,
			},
		},
		[]byte(`{"time":1654459632263}`),
		[][]byte{
			[]byte(`{"time":"2022-06-05T20:07:12.263000Z"}`),
		},
		nil,
	},
	{
		"JSON to unix_milli",
		config.Config{
			Type: "proc_time",
			Settings: map[string]interface{}{
				"key":        "time",
				"set_key":    "time",
				"format":     setFmt,
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
		"JSON unix to unix_milli",
		config.Config{
			Type: "proc_time",
			Settings: map[string]interface{}{
				"key":        "time",
				"set_key":    "time",
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
		"JSON offset conversion",
		config.Config{
			Type: "proc_time",
			Settings: map[string]interface{}{
				"key":        "time",
				"set_key":    "time",
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
		"JSON offset to local conversion",
		config.Config{
			Type: "proc_time",
			Settings: map[string]interface{}{
				"key":          "time",
				"set_key":      "time",
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
		"JSON local to local conversion",
		config.Config{
			Type: "proc_time",
			Settings: map[string]interface{}{
				"key":          "time",
				"set_key":      "time",
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

func TestProcTime(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procTimeTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcTime(ctx, test.cfg)
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

func benchmarkProcTime(b *testing.B, tform *procTime, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcTime(b *testing.B) {
	for _, test := range procTimeTests {
		proc, err := newProcTime(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcTime(b, proc, test.test)
			},
		)
	}
}
