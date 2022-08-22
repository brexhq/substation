package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
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
	cap := config.NewCapsule()
	for _, test := range captureTests {
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

func benchmarkCapture(b *testing.B, applicator Capture, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkCapture(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range captureTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkCapture(b, test.proc, cap)
			},
		)
	}
}
