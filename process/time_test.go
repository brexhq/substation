package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var outputFmt = "2006-01-02T15:04:05.000000Z"

var timeTests = []struct {
	name     string
	proc     Time
	err      error
	test     []byte
	expected []byte
}{
	{
		"data string",
		Time{
			Options: TimeOptions{
				InputFormat:  "2006-01-02T15:04:05Z",
				OutputFormat: outputFmt,
			},
		},
		nil,
		[]byte(`2021-03-06T00:02:57Z`),
		[]byte(`2021-03-06T00:02:57.000000Z`),
	},
	{
		"data unix",
		Time{
			Options: TimeOptions{
				InputFormat:  "unix",
				OutputFormat: outputFmt,
			},
		},
		nil,
		[]byte(`1639877490.061`),
		[]byte(`2021-12-19T01:31:30.061000Z`),
	},
	{
		"data unix to unix_milli",
		Time{
			Options: TimeOptions{
				InputFormat:  "unix",
				OutputFormat: "unix_milli",
			},
		},
		nil,
		[]byte(`1639877490.061`),
		[]byte(`1639877490061`),
	},
	{
		"string",
		Time{
			Options: TimeOptions{
				InputFormat:  "2006-01-02T15:04:05Z",
				OutputFormat: outputFmt,
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		nil,
		[]byte(`{"time":"2021-03-06T00:02:57Z"}`),
		[]byte(`{"time":"2021-03-06T00:02:57.000000Z"}`),
	},
	{
		"from unix",
		Time{
			Options: TimeOptions{
				InputFormat:  "unix",
				OutputFormat: outputFmt,
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		nil,
		[]byte(`{"time":1639877490}`),
		[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
	},
	{
		"to unix",
		Time{
			Options: TimeOptions{
				InputFormat:  outputFmt,
				OutputFormat: "unix",
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		nil,
		[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
		[]byte(`{"time":1639877490}`),
	},
	{
		"from unix_milli",
		Time{
			Options: TimeOptions{
				InputFormat:  "unix_milli",
				OutputFormat: outputFmt,
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		nil,
		[]byte(`{"time":1654459632263}`),
		[]byte(`{"time":"2022-06-05T20:07:12.263000Z"}`),
	},
	{
		"to unix_milli",
		Time{
			Options: TimeOptions{
				InputFormat:  outputFmt,
				OutputFormat: "unix_milli",
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		nil,
		[]byte(`{"time":"2022-06-05T20:07:12.263000Z"}`),
		[]byte(`{"time":1654459632263}`),
	},
	{
		"unix to unix_milli",
		Time{
			Options: TimeOptions{
				InputFormat:  "unix",
				OutputFormat: "unix_milli",
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		nil,
		[]byte(`{"time":1639877490}`),
		[]byte(`{"time":1639877490000}`),
	},
	{
		"offset conversion",
		Time{
			Options: TimeOptions{
				InputFormat:  "2006-Jan-02 Monday 03:04:05 -0700",
				OutputFormat: "2006-Jan-02 Monday 03:04:05 -0700",
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		nil,
		[]byte(`{"time":"2020-Jan-29 Wednesday 12:19:25 -0500"}`),
		[]byte(`{"time":"2020-Jan-29 Wednesday 05:19:25 +0000"}`),
	},
	{
		"offset to local conversion",
		Time{
			Options: TimeOptions{
				InputFormat:    "2006-Jan-02 Monday 03:04:05 -0700",
				OutputFormat:   "2006-Jan-02 Monday 03:04:05 PM",
				OutputLocation: "America/New_York",
			},
			InputKey:  "time",
			OutputKey: "time",
		},
		nil,
		// 12:19:25 AM in Pacific Standard Time
		[]byte(`{"time":"2020-Jan-29 Wednesday 00:19:25 -0800"}`),
		// 03:19:25 AM in Eastern Standard Time
		[]byte(`{"time":"2020-Jan-29 Wednesday 03:19:25 AM"}`),
	},
	{
		"local to local conversion",
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
		nil,
		// 12:19:25 AM in Pacific Standard Time
		[]byte(`{"time":"2020-Jan-29 Wednesday 00:19:25"}`),
		// 03:19:25 AM in Eastern Standard Time
		[]byte(`{"time":"2020-Jan-29 Wednesday 03:19:25"}`),
	},
	{
		"missing required options",
		Time{
			Options: TimeOptions{},
		},
		ProcessorInvalidSettings,
		[]byte{},
		[]byte{},
	},
}

func TestTime(t *testing.T) {
	ctx := context.TODO()
	for _, test := range timeTests {
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.As(err, &test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkTimeByte(b *testing.B, byter Time, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkTimeByte(b *testing.B) {
	for _, test := range timeTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkTimeByte(b, test.proc, test.test)
			},
		)
	}
}
