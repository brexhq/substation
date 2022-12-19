package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var outputFmt = "2006-01-02T15:04:05.000000Z"

var timeTests = []struct {
	name     string
	proc     time
	test     []byte
	expected []byte
	err      error
}{
	{
		"data string",
		time{
			Options: timeOptions{
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
		time{
			Options: timeOptions{
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
		time{
			Options: timeOptions{
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
		time{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: timeOptions{
				InputFormat:  "2006-01-02T15:04:05Z",
				OutputFormat: outputFmt,
			},
		},
		[]byte(`{"foo":"2021-03-06T00:02:57Z"}`),
		[]byte(`{"foo":"2021-03-06T00:02:57.000000Z"}`),
		nil,
	},
	{
		"JSON from unix",
		time{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: timeOptions{
				InputFormat:  "unix",
				OutputFormat: outputFmt,
			},
		},
		[]byte(`{"foo":1639877490}`),
		[]byte(`{"foo":"2021-12-19T01:31:30.000000Z"}`),
		nil,
	},
	{
		"JSON to unix",
		time{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: timeOptions{
				InputFormat:  outputFmt,
				OutputFormat: "unix",
			},
		},
		[]byte(`{"foo":"2021-12-19T01:31:30.000000Z"}`),
		[]byte(`{"foo":1639877490}`),
		nil,
	},
	{
		"JSON from unix_milli",
		time{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: timeOptions{
				InputFormat:  "unix_milli",
				OutputFormat: outputFmt,
			},
		},
		[]byte(`{"foo":1654459632263}`),
		[]byte(`{"foo":"2022-06-05T20:07:12.263000Z"}`),
		nil,
	},
	{
		"JSON to unix_milli",
		time{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: timeOptions{
				InputFormat:  outputFmt,
				OutputFormat: "unix_milli",
			},
		},
		[]byte(`{"foo":"2022-06-05T20:07:12.263000Z"}`),
		[]byte(`{"foo":1654459632263}`),
		nil,
	},
	{
		"JSON unix to unix_milli",
		time{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: timeOptions{
				InputFormat:  "unix",
				OutputFormat: "unix_milli",
			},
		},
		[]byte(`{"foo":1639877490}`),
		[]byte(`{"foo":1639877490000}`),
		nil,
	},
	{
		"JSON offset conversion",
		time{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: timeOptions{
				InputFormat:  "2006-Jan-02 Monday 03:04:05 -0700",
				OutputFormat: "2006-Jan-02 Monday 03:04:05 -0700",
			},
		},
		[]byte(`{"foo":"2020-Jan-29 Wednesday 12:19:25 -0500"}`),
		[]byte(`{"foo":"2020-Jan-29 Wednesday 05:19:25 +0000"}`),
		nil,
	},
	{
		"JSON offset to local conversion",
		time{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: timeOptions{
				InputFormat:    "2006-Jan-02 Monday 03:04:05 -0700",
				OutputFormat:   "2006-Jan-02 Monday 03:04:05 PM",
				OutputLocation: "America/New_York",
			},
		},
		// 12:19:25 AM in Pacific Standard time
		[]byte(`{"foo":"2020-Jan-29 Wednesday 00:19:25 -0800"}`),
		// 03:19:25 AM in Eastern Standard time
		[]byte(`{"foo":"2020-Jan-29 Wednesday 03:19:25 AM"}`),
		nil,
	},
	{
		"JSON local to local conversion",
		time{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: timeOptions{
				InputFormat:    "2006-Jan-02 Monday 03:04:05",
				OutputFormat:   "2006-Jan-02 Monday 03:04:05",
				InputLocation:  "America/Los_Angeles",
				OutputLocation: "America/New_York",
			},
		},
		// 12:19:25 AM in Pacific Standard time
		[]byte(`{"foo":"2020-Jan-29 Wednesday 00:19:25"}`),
		// 03:19:25 AM in Eastern Standard time
		[]byte(`{"foo":"2020-Jan-29 Wednesday 03:19:25"}`),
		nil,
	},
}

func TestTime(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range timeTests {
		capsule.SetData(test.test)

		result, err := test.proc.Apply(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkTime(b *testing.B, applicator time, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkTime(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range timeTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkTime(b, test.proc, capsule)
			},
		)
	}
}
