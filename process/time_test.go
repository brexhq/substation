package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
)

var outputFmt = "2006-01-02T15:04:05.000000Z"

var timeTests = []struct {
	name     string
	proc     Time
	test     []byte
	expected []byte
	err      error
}{
	{
		"data string",
		Time{
			Options: TimeOptions{
				InputFormat:  "2006-01-02T15:04:05Z",
				OutputFormat: outputFmt,
			},
		},
		[]byte(`2021-03-06T00:02:57Z`),
		[]byte(`2021-03-06T00:02:57.000000Z`),
		nil,
	},
	{
		"data unix",
		Time{
			Options: TimeOptions{
				InputFormat:  "unix",
				OutputFormat: outputFmt,
			},
		},
		[]byte(`1639877490.061`),
		[]byte(`2021-12-19T01:31:30.061000Z`),
		nil,
	},
	{
		"data unix to unix_milli",
		Time{
			Options: TimeOptions{
				InputFormat:  "unix",
				OutputFormat: "unix_milli",
			},
		},
		[]byte(`1639877490.061`),
		[]byte(`1639877490061`),
		nil,
	},
	{
		"JSON",
		Time{
			Options: TimeOptions{
				InputFormat:  "2006-01-02T15:04:05Z",
				OutputFormat: outputFmt,
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		[]byte(`{"time":"2021-03-06T00:02:57Z"}`),
		[]byte(`{"time":"2021-03-06T00:02:57.000000Z"}`),
		nil,
	},
	{
		"JSON from unix",
		Time{
			Options: TimeOptions{
				InputFormat:  "unix",
				OutputFormat: outputFmt,
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		[]byte(`{"time":1639877490}`),
		[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
		nil,
	},
	{
		"JSON to unix",
		Time{
			Options: TimeOptions{
				InputFormat:  outputFmt,
				OutputFormat: "unix",
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
		[]byte(`{"time":1639877490}`),
		nil,
	},
	{
		"JSON from unix_milli",
		Time{
			Options: TimeOptions{
				InputFormat:  "unix_milli",
				OutputFormat: outputFmt,
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		[]byte(`{"time":1654459632263}`),
		[]byte(`{"time":"2022-06-05T20:07:12.263000Z"}`),
		nil,
	},
	{
		"JSON to unix_milli",
		Time{
			Options: TimeOptions{
				InputFormat:  outputFmt,
				OutputFormat: "unix_milli",
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		[]byte(`{"time":"2022-06-05T20:07:12.263000Z"}`),
		[]byte(`{"time":1654459632263}`),
		nil,
	},
	{
		"JSON unix to unix_milli",
		Time{
			Options: TimeOptions{
				InputFormat:  "unix",
				OutputFormat: "unix_milli",
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		[]byte(`{"time":1639877490}`),
		[]byte(`{"time":1639877490000}`),
		nil,
	},
	{
		"JSON offset conversion",
		Time{
			Options: TimeOptions{
				InputFormat:  "2006-Jan-02 Monday 03:04:05 -0700",
				OutputFormat: "2006-Jan-02 Monday 03:04:05 -0700",
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		[]byte(`{"time":"2020-Jan-29 Wednesday 12:19:25 -0500"}`),
		[]byte(`{"time":"2020-Jan-29 Wednesday 05:19:25 +0000"}`),
		nil,
	},
	{
		"JSON offset to local conversion",
		Time{
			Options: TimeOptions{
				InputFormat:    "2006-Jan-02 Monday 03:04:05 -0700",
				OutputFormat:   "2006-Jan-02 Monday 03:04:05 PM",
				OutputLocation: "America/New_York",
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		// 12:19:25 AM in Pacific Standard Time
		[]byte(`{"time":"2020-Jan-29 Wednesday 00:19:25 -0800"}`),
		// 03:19:25 AM in Eastern Standard Time
		[]byte(`{"time":"2020-Jan-29 Wednesday 03:19:25 AM"}`),
		nil,
	},
	{
		"JSON local to local conversion",
		Time{
			Options: TimeOptions{
				InputFormat:    "2006-Jan-02 Monday 03:04:05",
				OutputFormat:   "2006-Jan-02 Monday 03:04:05",
				InputLocation:  "America/Los_Angeles",
				OutputLocation: "America/New_York",
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		// 12:19:25 AM in Pacific Standard Time
		[]byte(`{"time":"2020-Jan-29 Wednesday 00:19:25"}`),
		// 03:19:25 AM in Eastern Standard Time
		[]byte(`{"time":"2020-Jan-29 Wednesday 03:19:25"}`),
		nil,
	},
	{
		"invalid settings",
		Time{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
	},
}

func TestTime(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range timeTests {
		cap.SetData(test.test)

		res, err := test.proc.Apply(ctx, cap)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res.GetData(), test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res.GetData())
			t.Fail()
		}
	}
}

func benchmarkTime(b *testing.B, applicator Time, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkTime(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range timeTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkTime(b, test.proc, cap)
			},
		)
	}
}
