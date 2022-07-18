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
		"JSON find",
		Capture{
			Options: CaptureOptions{
				Expression: "^([^@]*)@.*$",
				Function:   "find",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"bar@qux.corge"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"JSON find_all",
		Capture{
			Options: CaptureOptions{
				Expression: "(.{1})",
				Function:   "find_all",
				Count:      3,
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":["b","a","r"]}`),
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
		[]byte(`bar@qux.corge`),
		[]byte(`bar`),
		nil,
	},
	{
		"named_group",
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
	ctx := context.TODO()
	for _, test := range captureTests {
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.Is(err, test.err) {
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
