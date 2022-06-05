package process

import (
	"bytes"
	"context"
	"testing"
)

var outputFmt = "2006-01-02T15:04:05.000000Z"

var timeTests = []struct {
	name     string
	proc     Time
	test     []byte
	expected []byte
}{
	{
		"string",
		Time{
			Input: Input{
				Key: "time",
			},
			Options: TimeOptions{
				InputFormat:  "2006-01-02T15:04:05Z",
				OutputFormat: outputFmt,
			},
			Output: Output{
				Key: "time",
			},
		},
		[]byte(`{"time":"2021-03-06T00:02:57Z"}`),
		[]byte(`{"time":"2021-03-06T00:02:57.000000Z"}`),
	},
	{
		"from unix",
		Time{
			Input: Input{
				Key: "time",
			},
			Options: TimeOptions{
				InputFormat:  "unix",
				OutputFormat: outputFmt,
			},
			Output: Output{
				Key: "time",
			},
		},
		[]byte(`{"time":1639877490}`),
		[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
	},
	{
		"to unix",
		Time{
			Input: Input{
				Key: "time",
			},
			Options: TimeOptions{
				InputFormat:  outputFmt,
				OutputFormat: "unix",
			},
			Output: Output{
				Key: "time",
			},
		},
		[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
		[]byte(`{"time":1639877490}`),
	},
	{
		"from unix_milli",
		Time{
			Input: Input{
				Key: "time",
			},
			Options: TimeOptions{
				InputFormat:  "unix_milli",
				OutputFormat: outputFmt,
			},
			Output: Output{
				Key: "time",
			},
		},
		[]byte(`{"time":1654459632263}`),
		[]byte(`{"time":"2022-06-05T20:07:12.263000Z"}`),
	},
	{
		"to unix_milli",
		Time{
			Input: Input{
				Key: "time",
			},
			Options: TimeOptions{
				InputFormat:  outputFmt,
				OutputFormat: "unix_milli",
			},
			Output: Output{
				Key: "time",
			},
		},
		[]byte(`{"time":"2022-06-05T20:07:12.263000Z"}`),
		[]byte(`{"time":1654459632263}`),
	},
	{
		"offset conversion",
		Time{
			Input: Input{
				Key: "time",
			},
			Options: TimeOptions{
				InputFormat:  "2006-Jan-02 Monday 03:04:05 -0700",
				OutputFormat: "2006-Jan-02 Monday 03:04:05 -0700",
			},
			Output: Output{
				Key: "time",
			},
		},
		[]byte(`{"time":"2020-Jan-29 Wednesday 12:19:25 -0500"}`),
		[]byte(`{"time":"2020-Jan-29 Wednesday 05:19:25 +0000"}`),
	},
	{
		"offset to local conversion",
		Time{
			Input: Input{
				Key: "time",
			},
			Options: TimeOptions{
				InputFormat:    "2006-Jan-02 Monday 03:04:05 -0700",
				OutputFormat:   "2006-Jan-02 Monday 03:04:05 PM",
				OutputLocation: "America/New_York",
			},
			Output: Output{
				Key: "time",
			},
		},
		// 12:19:25 AM in Pacific Standard Time
		[]byte(`{"time":"2020-Jan-29 Wednesday 00:19:25 -0800"}`),
		// 03:19:25 AM in Eastern Standard Time
		[]byte(`{"time":"2020-Jan-29 Wednesday 03:19:25 AM"}`),
	},
	{
		"local to local conversion",
		Time{
			Input: Input{
				Key: "time",
			},
			Options: TimeOptions{
				InputFormat:    "2006-Jan-02 Monday 03:04:05",
				InputLocation:  "America/Los_Angeles",
				OutputFormat:   "2006-Jan-02 Monday 03:04:05",
				OutputLocation: "America/New_York",
			},
			Output: Output{
				Key: "time",
			},
		},
		// 12:19:25 AM in Pacific Standard Time
		[]byte(`{"time":"2020-Jan-29 Wednesday 00:19:25"}`),
		// 03:19:25 AM in Eastern Standard Time
		[]byte(`{"time":"2020-Jan-29 Wednesday 03:19:25"}`),
	},
	{
		"array",
		Time{
			Input: Input{
				Key: "time",
			},
			Options: TimeOptions{
				InputFormat:  "2006-01-02T15:04:05Z",
				OutputFormat: outputFmt,
			},
			Output: Output{
				Key: "time",
			},
		},
		[]byte(`{"time":["2021-03-06T00:02:57Z","2021-03-06T00:03:57Z"]}`),
		[]byte(`{"time":["2021-03-06T00:02:57.000000Z","2021-03-06T00:03:57.000000Z"]}`),
	},
	{
		"array",
		Time{
			Input: Input{
				Key: "time",
			},
			Options: TimeOptions{
				InputFormat:  "unix",
				OutputFormat: outputFmt,
			},
			Output: Output{
				Key: "time",
			},
		},
		[]byte(`{"time":[1639877490.061,1651705967]}`),
		[]byte(`{"time":["2021-12-19T01:31:30.000000Z","2022-05-04T23:12:47.000000Z"]}`),
	},
}

func TestTime(t *testing.T) {
	ctx := context.TODO()
	for _, test := range timeTests {
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil {
			t.Logf("%v", err)
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
