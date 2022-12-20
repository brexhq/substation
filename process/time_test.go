package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var setFmt = "2006-01-02T15:04:05.000000Z"

var timeTests = []struct {
	name     string
	proc     _time
	test     []byte
	expected []byte
	err      error
}{
	{
		"data string",
		_time{
			Options: _timeOptions{
				Format:    "2006-01-02T15:04:05Z",
				SetFormat: setFmt,
			},
		},
		[]byte(`2021-03-06T00:02:57Z`),
		[]byte(`2021-03-06T00:02:57.000000Z`),
		nil,
	},
	{
		"data unix",
		_time{
			Options: _timeOptions{
				Format:    "unix",
				SetFormat: setFmt,
			},
		},
		[]byte(`1639877490.061`),
		[]byte(`2021-12-19T01:31:30.061000Z`),
		nil,
	},
	{
		"data unix to unix_milli",
		_time{
			Options: _timeOptions{
				Format:    "unix",
				SetFormat: "unix_milli",
			},
		},
		[]byte(`1639877490.061`),
		[]byte(`1639877490061`),
		nil,
	},
	{
		"JSON",
		_time{
			process: process{
				Key:    "time",
				SetKey: "time",
			},
			Options: _timeOptions{
				Format:    "2006-01-02T15:04:05Z",
				SetFormat: setFmt,
			},
		},
		[]byte(`{"time":"2021-03-06T00:02:57Z"}`),
		[]byte(`{"time":"2021-03-06T00:02:57.000000Z"}`),
		nil,
	},
	{
		"JSON from unix",
		_time{
			process: process{
				Key:    "time",
				SetKey: "time",
			},
			Options: _timeOptions{
				Format:    "unix",
				SetFormat: setFmt,
			},
		},
		[]byte(`{"time":1639877490}`),
		[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
		nil,
	},
	{
		"JSON to unix",
		_time{
			process: process{
				Key:    "time",
				SetKey: "time",
			},
			Options: _timeOptions{
				Format:    setFmt,
				SetFormat: "unix",
			},
		},
		[]byte(`{"time":"2021-12-19T01:31:30.000000Z"}`),
		[]byte(`{"time":1639877490}`),
		nil,
	},
	{
		"JSON from unix_milli",
		_time{
			process: process{
				Key:    "time",
				SetKey: "time",
			},
			Options: _timeOptions{
				Format:    "unix_milli",
				SetFormat: setFmt,
			},
		},
		[]byte(`{"time":1654459632263}`),
		[]byte(`{"time":"2022-06-05T20:07:12.263000Z"}`),
		nil,
	},
	{
		"JSON to unix_milli",
		_time{
			process: process{
				Key:    "time",
				SetKey: "time",
			},
			Options: _timeOptions{
				Format:    setFmt,
				SetFormat: "unix_milli",
			},
		},
		[]byte(`{"time":"2022-06-05T20:07:12.263000Z"}`),
		[]byte(`{"time":1654459632263}`),
		nil,
	},
	{
		"JSON unix to unix_milli",
		_time{
			process: process{
				Key:    "time",
				SetKey: "time",
			},
			Options: _timeOptions{
				Format:    "unix",
				SetFormat: "unix_milli",
			},
		},
		[]byte(`{"time":1639877490}`),
		[]byte(`{"time":1639877490000}`),
		nil,
	},
	{
		"JSON offset conversion",
		_time{
			process: process{
				Key:    "time",
				SetKey: "time",
			},
			Options: _timeOptions{
				Format:    "2006-Jan-02 Monday 03:04:05 -0700",
				SetFormat: "2006-Jan-02 Monday 03:04:05 -0700",
			},
		},
		[]byte(`{"time":"2020-Jan-29 Wednesday 12:19:25 -0500"}`),
		[]byte(`{"time":"2020-Jan-29 Wednesday 05:19:25 +0000"}`),
		nil,
	},
	{
		"JSON offset to local conversion",
		_time{
			process: process{
				Key:    "time",
				SetKey: "time",
			},
			Options: _timeOptions{
				Format:      "2006-Jan-02 Monday 03:04:05 -0700",
				SetFormat:   "2006-Jan-02 Monday 03:04:05 PM",
				SetLocation: "America/New_York",
			},
		},
		// 12:19:25 AM in Pacific Standard time
		[]byte(`{"time":"2020-Jan-29 Wednesday 00:19:25 -0800"}`),
		// 03:19:25 AM in Eastern Standard time
		[]byte(`{"time":"2020-Jan-29 Wednesday 03:19:25 AM"}`),
		nil,
	},
	{
		"JSON local to local conversion",
		_time{
			process: process{
				Key:    "time",
				SetKey: "time",
			},
			Options: _timeOptions{
				Format:      "2006-Jan-02 Monday 03:04:05",
				Location:    "America/Los_Angeles",
				SetFormat:   "2006-Jan-02 Monday 03:04:05",
				SetLocation: "America/New_York",
			},
		},
		// 12:19:25 AM in Pacific Standard time
		[]byte(`{"time":"2020-Jan-29 Wednesday 00:19:25"}`),
		// 03:19:25 AM in Eastern Standard time
		[]byte(`{"time":"2020-Jan-29 Wednesday 03:19:25"}`),
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

func benchmarkTime(b *testing.B, applicator _time, test config.Capsule) {
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
