package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var captureTests = []struct {
	name     string
	proc     Capture
	test     []byte
	expected []byte
	err      error
}{
	{
		"json find",
		Capture{
			InputKey:  "capture",
			OutputKey: "capture",
			Options: CaptureOptions{
				Expression: "^([^@]*)@.*$",
				Function:   "find",
			},
		},
		[]byte(`{"capture":"foo@qux.com"}`),
		[]byte(`{"capture":"foo"}`),
		nil,
	},
	{
		"json find_all",
		Capture{
			InputKey:  "capture",
			OutputKey: "capture",
			Options: CaptureOptions{
				Expression: "(.{1})",
				Function:   "find_all",
				Count:      3,
			},
		},
		[]byte(`{"capture":"foo"}`),
		[]byte(`{"capture":["f","o","o"]}`),
		nil,
	},
	{
		"data",
		Capture{
			Options: CaptureOptions{
				Expression: "^([^@]*)@.*$",
				Function:   "find",
			},
		},
		[]byte(`foo@qux.com`),
		[]byte(`foo`),
		nil,
	},
	{
		"data",
		Capture{
			Options: CaptureOptions{
				Function:   "named_group",
				Expression: "(?P<foo>[a-zA-Z]+) (?P<qux>[a-zA-Z]+)",
			},
		},
		[]byte(`bar quux`),
		[]byte(`{"foo":"bar","qux":"quux"}`),
		nil,
	},
	{
		"invalid settings",
		Capture{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
	},
}

func TestCapture(t *testing.T) {
	for _, test := range captureTests {
		ctx := context.TODO()
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

func benchmarkCaptureByte(b *testing.B, byter Capture, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkCaptureByte(b *testing.B) {
	for _, test := range captureTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkCaptureByte(b, test.proc, test.test)
			},
		)
	}
}
