package process

import (
	"bytes"
	"context"
	"testing"
)

var outputFmt = "2006-01-02T15:04:05.000000Z"

func TestTime(t *testing.T) {
	var tests = []struct {
		proc     Time
		test     []byte
		expected []byte
	}{
		{
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
			[]byte(`{"time":1639877490.061}`),
			[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
		},
		{
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
			[]byte(`{"time":1639877490.061}`),
			[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
		},
		// offset conversion
		{
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
		// offset to local conversion
		{
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
		// local to local conversion
		{
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
		// array input
		{
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
	}

	for _, test := range tests {
		ctx := context.TODO()
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
